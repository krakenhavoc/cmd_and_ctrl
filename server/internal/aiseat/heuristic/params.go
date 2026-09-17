package heuristic

import "encoding/json"

// params.go re-declares the wire payload shapes `internal/legal`
// emits, so the policy can read a Move's parameters without a handle
// on the engine. They are deliberately duplicated rather than
// imported: legal's param structs are unexported, and exporting them
// would make the policy depend on the enumerator's internals when
// what it actually depends on is the WIRE, which is stable.
//
// Only the fields the policy reasons about are declared; the rest of
// each payload (strict, auto_tap, face, …) decodes and is dropped.
//
// paramsFor* return the zero value on a payload that does not parse.
// That is not a silent failure worth shouting about: a move whose
// params the policy cannot read simply scores as an unknown, and the
// enumerator's own tests are what guarantee the shapes.

type targetRef struct {
	Kind string `json:"kind"`
	ID   string `json:"id"`
}

type castParams struct {
	InstanceID   string      `json:"instance_id"`
	FromZone     string      `json:"from_zone"`
	Targets      []targetRef `json:"targets"`
	Modes        []int       `json:"modes"`
	XValue       int         `json:"x_value"`
	DiscardIDs   []string    `json:"discard_ids"`
	SacrificeIDs []string    `json:"sacrifice_ids"`
}

type activateParams struct {
	SourceCardID string      `json:"source_card_id"`
	AbilityIndex int         `json:"ability_index"`
	Targets      []targetRef `json:"targets"`
	SacrificeIDs []string    `json:"sacrifice_ids"`
	CrewIDs      []string    `json:"crew_ids"`
	XValue       int         `json:"x_value"`
}

type attackParams struct {
	Attacker string `json:"attacker"`
	Target   string `json:"target"`
}

type blockParams struct {
	Blocker  string `json:"blocker"`
	Attacker string `json:"attacker"`
}

// choiceParams covers every resolve_choice shape. The dispatcher
// routes on which field is present and so does the policy, except
// that the policy also has the choice's Kind available from
// GameView.PendingChoices, which is a far better signal than
// guessing from the populated fields.
type choiceParams struct {
	ChoiceID string      `json:"choice_id"`
	CardIDs  []string    `json:"card_ids"`
	Color    string      `json:"color"`
	Order    []string    `json:"order"`
	Apply    *bool       `json:"apply"`
	Target   *targetRef  `json:"target"`
	Targets  []targetRef `json:"targets"`
	Bottom   []string    `json:"bottom"`
	TopOrder []string    `json:"top_order"`
	Call     string      `json:"call"`
}

type mulliganParams struct {
	HandSize int `json:"hand_size"`
}

type discardSelectionParams struct {
	CardIDs []string `json:"card_ids"`
}

func decode[T any](raw json.RawMessage) T {
	var out T
	if len(raw) > 0 {
		_ = json.Unmarshal(raw, &out)
	}
	return out
}
