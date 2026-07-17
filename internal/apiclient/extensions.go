package apiclient

import (
	"bytes"
	"encoding/json"
	"strconv"
)

// UnmarshalJSON tolerates the two shapes the Anthropic API returns for an
// agent's model field: either a bare string (older `agent-api-*` betas) or
// an object `{"id": "...", "speed": "..."}` (newer `managed-agents-*` betas).
func (a *AgentModelConfig) UnmarshalJSON(data []byte) error {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) > 0 && trimmed[0] == '"' {
		var s string
		if err := json.Unmarshal(trimmed, &s); err != nil {
			return err
		}
		a.Id = s
		a.Speed = nil
		return nil
	}
	type raw AgentModelConfig
	var r raw
	if err := json.Unmarshal(trimmed, &r); err != nil {
		return err
	}
	*a = AgentModelConfig(r)
	return nil
}

// UnmarshalJSON tolerates `version` arriving as either a JSON string
// ("18", what `agent-api-2026-03-01` returns) or a JSON number (18, what
// `managed-agents-2026-04-01` returns) so the same struct works against
// either beta header.
func (a *Agent) UnmarshalJSON(data []byte) error {
	type raw Agent
	// First decode everything except `version`, which we handle out of band.
	var probe map[string]json.RawMessage
	if err := json.Unmarshal(data, &probe); err != nil {
		return err
	}
	versionRaw, hadVersion := probe["version"]
	if hadVersion {
		// Strip version so the standard decode won't fight over its type.
		delete(probe, "version")
	}
	cleaned, err := json.Marshal(probe)
	if err != nil {
		return err
	}
	var r raw
	if err := json.Unmarshal(cleaned, &r); err != nil {
		return err
	}
	*a = Agent(r)
	if hadVersion {
		v, err := stringFromStringOrNumber(versionRaw)
		if err != nil {
			return err
		}
		a.Version = v
	}
	return nil
}

// UnmarshalJSON tolerates `version` arriving as either a JSON string or a JSON
// number in a deployment's embedded agent reference — same inconsistency as
// Agent.version above (managed-agents-2026-04-01 returns it as a number).
func (a *DeploymentAgentRef) UnmarshalJSON(data []byte) error {
	type raw DeploymentAgentRef
	var probe map[string]json.RawMessage
	if err := json.Unmarshal(data, &probe); err != nil {
		return err
	}
	versionRaw, hadVersion := probe["version"]
	if hadVersion {
		delete(probe, "version")
	}
	cleaned, err := json.Marshal(probe)
	if err != nil {
		return err
	}
	var r raw
	if err := json.Unmarshal(cleaned, &r); err != nil {
		return err
	}
	*a = DeploymentAgentRef(r)
	if hadVersion {
		v, err := stringFromStringOrNumber(versionRaw)
		if err != nil {
			return err
		}
		a.Version = v
	}
	return nil
}

func stringFromStringOrNumber(data json.RawMessage) (string, error) {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 {
		return "", nil
	}
	if trimmed[0] == '"' {
		var s string
		if err := json.Unmarshal(trimmed, &s); err != nil {
			return "", err
		}
		return s, nil
	}
	// Number — preserve the literal so int-vs-float doesn't reformat it.
	var n json.Number
	dec := json.NewDecoder(bytes.NewReader(trimmed))
	dec.UseNumber()
	if err := dec.Decode(&n); err != nil {
		return "", err
	}
	// Normalize floats like 18.0 to "18".
	if i, err := n.Int64(); err == nil {
		return strconv.FormatInt(i, 10), nil
	}
	return n.String(), nil
}
