// Copyright (c) 2026 John Dewey

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to
// deal in the Software without restriction, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense, and/or
// sell copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER
// DEALINGS IN THE SOFTWARE.

package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/retr0h/mlb-sdk/pkg/api"
)

// composedOperationIDs is the set of OpenAPI operationIds that already have a
// hand-written composed tool. registerSpecTools skips these so the raw mlb_*
// tools don't shadow the richer versions.
var composedOperationIDs = map[string]bool{
	"getSchedule":           true, // → scores
	"getStandings":          true, // → standings
	"getPerson":             true, // → player_bio
	"getTeam":               true, // → team_info
	"getTeamRoster":         true, // → team_roster
	"getStatsLeaders":       true, // → league_leaders
	"getBoxscore":           true, // → game_detail
	"getLinescore":          true, // → game_linescore
	"getTransactions":       true, // → recent_transactions
	"getFreeAgents":         true, // → free_agents
	"getSchedulePostseason": true, // → postseason_schedule
}

// specParam describes one parameter extracted from the OpenAPI spec.
type specParam struct {
	name        string // as in the spec, e.g. "gamePk"
	in          string // "path" | "query"
	description string
	required    bool
	typ         string // "integer" | "string" | "boolean" | "number"
}

// registerSpecTools parses the embedded OpenAPI spec and registers a raw
// mlb_<operationId> tool for every GET operation not already covered by a
// composed tool. Each handler makes a direct HTTP GET to statsapi.mlb.com
// and returns the raw JSON response body as a text result.
func (s *Server) registerSpecTools() {
	loader := openapi3.NewLoader()

	doc, err := loader.LoadFromData(api.OpenAPISpec())
	if err != nil {
		s.logger.Error("registerSpecTools: load spec", "err", err)
		return
	}

	// Collect and sort paths for deterministic registration order.
	paths := make([]string, 0, len(doc.Paths.Map()))
	for p := range doc.Paths.Map() {
		paths = append(paths, p)
	}
	sort.Strings(paths)

	for _, pathStr := range paths {
		item := doc.Paths.Map()[pathStr]
		if item == nil || item.Get == nil {
			continue
		}

		oper := item.Get
		if oper.OperationID == "" {
			continue
		}
		if composedOperationIDs[oper.OperationID] {
			continue
		}

		params := collectSpecParams(oper)

		toolName := "mlb_" + operationToSnake(oper.OperationID)
		desc := oper.Summary
		if oper.Description != "" {
			desc = oper.Summary + "\n\n" + strings.TrimSpace(oper.Description)
		}

		schema := buildInputSchema(params)

		// Use the raw Server.AddTool (not the generic mcpsdk.AddTool) because
		// the schema and parameter mapping are constructed dynamically here
		// rather than inferred from a typed Go struct.
		s.mcp.AddTool(
			&mcpsdk.Tool{
				Name:        toolName,
				Description: desc,
				InputSchema: schema,
			},
			s.makeSpecToolHandler(pathStr, params),
		)
	}
}

// collectSpecParams extracts path and query parameters from an operation.
func collectSpecParams(oper *openapi3.Operation) []specParam {
	params := make([]specParam, 0, len(oper.Parameters))
	for _, pRef := range oper.Parameters {
		p := pRef.Value
		if p == nil {
			continue
		}
		if p.In != "path" && p.In != "query" {
			continue
		}

		typ := "string"
		if p.Schema != nil && p.Schema.Value != nil {
			sv := p.Schema.Value
			if len(sv.Type.Slice()) > 0 {
				typ = sv.Type.Slice()[0]
			}
		}

		params = append(params, specParam{
			name:        p.Name,
			in:          p.In,
			description: p.Description,
			required:    p.Required,
			typ:         typ,
		})
	}
	return params
}

// buildInputSchema constructs a JSON Schema object (as map[string]any, which
// the MCP SDK accepts via its remarshal path) from a list of spec parameters.
func buildInputSchema(params []specParam) map[string]any {
	properties := make(map[string]any, len(params))
	var required []string

	// Sort for deterministic property ordering.
	sorted := make([]specParam, len(params))
	copy(sorted, params)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].name < sorted[j].name
	})

	for _, p := range sorted {
		prop := map[string]any{
			"type": specTypeToJSONSchemaType(p.typ),
		}
		if p.description != "" {
			prop["description"] = p.description
		}
		properties[p.name] = prop
		if p.required {
			required = append(required, p.name)
		}
	}

	schema := map[string]any{
		"type":       "object",
		"properties": properties,
	}
	if len(required) > 0 {
		schema["required"] = required
	}
	return schema
}

// specTypeToJSONSchemaType maps OpenAPI primitive type names to JSON Schema
// type names. Both vocabularies overlap almost entirely; "integer" is the one
// notable gap — JSON Schema treats it as a subtype of "number" but accepts the
// keyword literally.
func specTypeToJSONSchemaType(typ string) string {
	switch typ {
	case "integer":
		return "integer"
	case "number":
		return "number"
	case "boolean":
		return "boolean"
	default:
		return "string"
	}
}

// makeSpecToolHandler returns a raw ToolHandler that:
//  1. Unmarshals the MCP call arguments into a map[string]any.
//  2. Builds the statsapi.mlb.com URL from the OpenAPI path template + params.
//  3. Issues an HTTP GET with the caller's context.
//  4. Returns the raw JSON response body as a text result.
func (s *Server) makeSpecToolHandler(
	pathTemplate string,
	params []specParam,
) mcpsdk.ToolHandler {
	return func(
		ctx context.Context,
		req *mcpsdk.CallToolRequest,
	) (*mcpsdk.CallToolResult, error) {
		var args map[string]any
		if len(req.Params.Arguments) > 0 {
			if err := json.Unmarshal(req.Params.Arguments, &args); err != nil {
				return nil, fmt.Errorf("mlb-mcp: spec tool: unmarshal args: %w", err)
			}
		}

		rawURL, err := buildSpecURL("https://statsapi.mlb.com", pathTemplate, params, args)
		if err != nil {
			return nil, fmt.Errorf("mlb-mcp: spec tool: build url: %w", err)
		}

		httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
		if err != nil {
			return nil, fmt.Errorf("mlb-mcp: spec tool: new request: %w", err)
		}

		resp, err := http.DefaultClient.Do(httpReq)
		if err != nil {
			return nil, fmt.Errorf("mlb-mcp: spec tool: http get %s: %w", rawURL, err)
		}
		defer resp.Body.Close() //nolint:errcheck

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("mlb-mcp: spec tool: read body: %w", err)
		}

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return nil, fmt.Errorf(
				"mlb-mcp: spec tool: %s returned %s",
				rawURL,
				resp.Status,
			)
		}

		return textResult(string(body)), nil
	}
}

// buildSpecURL replaces path parameter placeholders in the template with
// values from args, then appends non-empty query parameters.
func buildSpecURL(
	base, pathTemplate string,
	params []specParam,
	args map[string]any,
) (string, error) {
	path := pathTemplate
	q := url.Values{}

	for _, p := range params {
		raw, ok := args[p.name]
		if !ok || raw == nil {
			if p.required {
				return "", fmt.Errorf("required parameter %q not provided", p.name)
			}
			continue
		}

		strVal := anyToString(raw)
		if strVal == "" {
			continue
		}

		switch p.in {
		case "path":
			path = strings.ReplaceAll(path, "{"+p.name+"}", url.PathEscape(strVal))
		case "query":
			q.Set(p.name, strVal)
		}
	}

	full := base + path
	if len(q) > 0 {
		full += "?" + q.Encode()
	}
	return full, nil
}

// anyToString converts a value from a map[string]any (MCP arguments, decoded
// from JSON) to a URL-parameter string. JSON numbers arrive as float64.
func anyToString(v any) string {
	switch x := v.(type) {
	case string:
		return x
	case float64:
		// Drop the decimal when the value is a whole number (119.0 → "119").
		if x == float64(int64(x)) {
			return fmt.Sprintf("%d", int64(x))
		}
		return fmt.Sprintf("%g", x)
	case bool:
		if x {
			return "true"
		}
		return "false"
	default:
		return fmt.Sprintf("%v", v)
	}
}

// operationToSnake converts a camelCase operationId (e.g. "getBoxscore") to
// snake_case (e.g. "get_boxscore") for use as a tool name suffix.
func operationToSnake(s string) string {
	var b strings.Builder
	for i, r := range s {
		if r >= 'A' && r <= 'Z' {
			if i > 0 {
				b.WriteByte('_')
			}
			b.WriteRune(r - 'A' + 'a')
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}
