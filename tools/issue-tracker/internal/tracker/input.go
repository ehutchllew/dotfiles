package tracker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/ehutchllew/dotfiles/tools/issue-tracker/internal/apperr"
	"github.com/ehutchllew/dotfiles/tools/issue-tracker/internal/doc"
)

// TicketRef is a ticket number given as a JSON string or number.
type TicketRef int

// UnmarshalJSON accepts "01", "1", or 1.
func (r *TicketRef) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		var n json.Number
		if err := json.Unmarshal(b, &n); err != nil {
			return fmt.Errorf("ticket must be a number such as \"01\"")
		}
		s = n.String()
	}
	n, err := doc.ParseTicketNumber(s)
	if err != nil {
		return err
	}
	*r = TicketRef(n)
	return nil
}

// decodeStrict decodes one JSON object into v. It rejects fields outside
// allowed, even when v has them, so each command accepts only its own fields.
func decodeStrict(data []byte, v any, allowed []string) error {
	if len(bytes.TrimSpace(data)) == 0 {
		return apperr.Usage("invalid_input", "input is empty; send a JSON object")
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return apperr.Usage("invalid_input", "input isn't a JSON object: %v", err)
	}
	ok := map[string]bool{}
	for _, a := range allowed {
		ok[a] = true
	}
	var bad []string
	for k := range raw {
		if !ok[k] {
			bad = append(bad, k)
		}
	}
	if len(bad) > 0 {
		sort.Strings(bad)
		return apperr.Usage("invalid_input", "unsupported input fields: %s", strings.Join(bad, ", ")).With("allowed", allowed)
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		return apperr.Usage("invalid_input", "invalid input: %v", err)
	}
	if _, err := dec.Token(); err != io.EOF {
		return apperr.Usage("invalid_input", "input has data after the JSON object")
	}
	return nil
}

// Resolution resolves one not-blocking finding.
type Resolution struct {
	Finding    string    `json:"finding"`
	Resolution string    `json:"resolution"`
	Ticket     TicketRef `json:"ticket"`
}

// OpInput is the payload for ticket state operations.
type OpInput struct {
	Commit         string       `json:"commit"`
	CriteriaMet    []string     `json:"criteria_met"`
	CriteriaUnmet  []string     `json:"criteria_unmet"`
	AllCriteriaMet bool         `json:"all_criteria_met"`
	Findings       []string     `json:"findings"`
	Verification   []string     `json:"verification"`
	Evidence       []string     `json:"evidence"`
	NotBlocking    []string     `json:"not_blocking"`
	Resolved       []Resolution `json:"resolved"`
	Reason         string       `json:"reason"`
	Notes          []string     `json:"notes"`
}

// SpecCreateInput is the payload for `spec create`.
type SpecCreateInput struct {
	Title                   string   `json:"title"`
	AttemptLimit            int      `json:"attempt_limit"`
	ProblemStatement        string   `json:"problem_statement"`
	Solution                string   `json:"solution"`
	UserStories             []string `json:"user_stories"`
	ImplementationDecisions string   `json:"implementation_decisions"`
	TestingDecisions        string   `json:"testing_decisions"`
	OutOfScope              string   `json:"out_of_scope"`
	Assumptions             string   `json:"assumptions"`
	FurtherNotes            string   `json:"further_notes"`
	History                 string   `json:"history"`
}

// SpecCreateFields lists the fields `spec create` accepts.
var SpecCreateFields = []string{"title", "attempt_limit", "problem_statement", "solution", "user_stories", "implementation_decisions", "testing_decisions", "out_of_scope", "assumptions", "further_notes", "history"}

// SpecEditInput is the payload for `spec edit`.
type SpecEditInput struct {
	Change                  string            `json:"change"`
	Approved                bool              `json:"approved"`
	History                 string            `json:"history"`
	Title                   *string           `json:"title"`
	AttemptLimit            *int              `json:"attempt_limit"`
	ProblemStatement        *string           `json:"problem_statement"`
	Solution                *string           `json:"solution"`
	ImplementationDecisions *string           `json:"implementation_decisions"`
	TestingDecisions        *string           `json:"testing_decisions"`
	OutOfScope              *string           `json:"out_of_scope"`
	Assumptions             *string           `json:"assumptions"`
	FurtherNotes            *string           `json:"further_notes"`
	AddStories              []string          `json:"add_stories"`
	EditStories             map[string]string `json:"edit_stories"`
	WithdrawStories         []string          `json:"withdraw_stories"`
}

// SpecEditFields lists the fields `spec edit` accepts.
var SpecEditFields = []string{"change", "approved", "history", "title", "attempt_limit", "problem_statement", "solution", "implementation_decisions", "testing_decisions", "out_of_scope", "assumptions", "further_notes", "add_stories", "edit_stories", "withdraw_stories"}

// TicketCreateInput is the payload for `ticket create`.
type TicketCreateInput struct {
	Title        string      `json:"title"`
	Kind         string      `json:"kind"`
	Slug         string      `json:"slug"`
	BlockedBy    []TicketRef `json:"blocked_by"`
	Covers       *[]string   `json:"covers"`
	SpecRevision int         `json:"spec_revision"`
	WhatToBuild  string      `json:"what_to_build"`
	Criteria     []string    `json:"criteria"`
	Verification string      `json:"verification"`
	OutOfScope   string      `json:"out_of_scope"`
}

// TicketCreateFields lists the fields `ticket create` accepts.
var TicketCreateFields = []string{"title", "kind", "slug", "blocked_by", "covers", "spec_revision", "what_to_build", "criteria", "verification", "out_of_scope"}

// TicketEditInput is the payload for `ticket edit`.
type TicketEditInput struct {
	Title        *string      `json:"title"`
	BlockedBy    *[]TicketRef `json:"blocked_by"`
	Covers       *[]string    `json:"covers"`
	SpecRevision *int         `json:"spec_revision"`
	WhatToBuild  *string      `json:"what_to_build"`
	Criteria     *[]string    `json:"criteria"`
	Verification *string      `json:"verification"`
	OutOfScope   *string      `json:"out_of_scope"`
}

// TicketEditFields lists the fields `ticket edit` accepts.
var TicketEditFields = []string{"title", "blocked_by", "covers", "spec_revision", "what_to_build", "criteria", "verification", "out_of_scope"}

// DecodeOp decodes a state operation payload.
func DecodeOp(op string, data []byte) (*OpInput, error) {
	spec, ok := opSpecs[op]
	if !ok {
		return nil, apperr.Usage("unknown_operation", "%q is not a ticket operation", op)
	}
	in := &OpInput{}
	if data == nil {
		return in, nil
	}
	if err := decodeStrict(data, in, spec.fields); err != nil {
		return nil, err
	}
	return in, nil
}

// DecodeSpecCreate decodes a `spec create` payload.
func DecodeSpecCreate(data []byte) (*SpecCreateInput, error) {
	in := &SpecCreateInput{}
	return in, decodeStrict(data, in, SpecCreateFields)
}

// DecodeSpecEdit decodes a `spec edit` payload.
func DecodeSpecEdit(data []byte) (*SpecEditInput, error) {
	in := &SpecEditInput{}
	return in, decodeStrict(data, in, SpecEditFields)
}

// DecodeTicketCreate decodes a `ticket create` payload.
func DecodeTicketCreate(data []byte) (*TicketCreateInput, error) {
	in := &TicketCreateInput{}
	return in, decodeStrict(data, in, TicketCreateFields)
}

// DecodeTicketEdit decodes a `ticket edit` payload.
func DecodeTicketEdit(data []byte) (*TicketEditInput, error) {
	in := &TicketEditInput{}
	return in, decodeStrict(data, in, TicketEditFields)
}
