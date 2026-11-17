package cli_test

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidation(t *testing.T) {
	spec := ".scratch/checkout/spec.md"
	cases := []struct {
		name   string
		setup  func(r *repo)
		want   []string // "code" or "ticket:code"
		absent []string
		full   bool // only full validation reports it
	}{
		{name: "missing blocker", setup: func(r *repo) { r.setMeta(co, "02", "Blocked by", "09") }, want: []string{"02:invalid_blocker"}},
		{name: "cancelled blocker", setup: func(r *repo) { r.setMeta(co, "03", "Status", "cancelled") }, want: []string{"04:invalid_blocker"}},
		{name: "self blocker", setup: func(r *repo) { r.setMeta(co, "01", "Blocked by", "01") }, want: []string{"01:invalid_blocker"}},
		{name: "cycle", setup: func(r *repo) { r.setMeta(co, "01", "Blocked by", "02") }, want: []string{"blocker_cycle"}},
		{name: "in-progress acceptance with open blockers", setup: func(r *repo) {
			r.setMeta(co, "04", "Status", "in-progress", "Base commit", shaA, "Target commit", shaB)
		}, want: []string{"04:blockers_not_done"}},
		{name: "unknown story", setup: func(r *repo) { r.setMeta(co, "01", "Covers", "US 9") }, want: []string{"01:unknown_story"}},
		{name: "withdrawn story", setup: func(r *repo) {
			r.edit(spec, "US 2: As a shopper, I want to see the count on my phone, so that I can check on the go", "US 2: Withdrawn in revision 1.")
			r.setMeta(co, "02", "Status", "done")
		}, want: []string{"03:withdrawn_story", "04:withdrawn_story"}, absent: []string{"02:withdrawn_story"}},
		{name: "bad phase", setup: func(r *repo) { r.setMeta(co, "01", "Phase", "testing") }, want: []string{"01:invalid_field"}},
		{name: "unknown field", setup: func(r *repo) { r.setMeta(co, "01", "Owner", "me") }, want: []string{"01:unknown_field"}},
		{name: "field for another kind", setup: func(r *repo) { r.setMeta(co, "01", "Spec revision", "1") }, want: []string{"01:field_not_allowed"}},
		{name: "missing field", setup: func(r *repo) { r.setMeta(co, "01", "Covers", "") }, want: []string{"01:missing_field"}},
		{name: "review in progress", setup: func(r *repo) {
			r.setMeta(co, "01", "Phase", "review", "Status", "in-progress", "Attempt", "1", "Base commit", shaA, "Target commit", shaB)
		}, want: []string{"01:invalid_state"}},
		{name: "attempt without base", setup: func(r *repo) { r.setMeta(co, "01", "Attempt", "1") }, want: []string{"01:invalid_state"}},
		{name: "short commit", setup: func(r *repo) { r.setMeta(co, "01", "Base commit", "abc123") }, want: []string{"01:invalid_field"}},
		{name: "no criteria", setup: func(r *repo) {
			p := r.ticket(co, "03").Path
			r.edit(p, "- [ ] The count is readable without zooming.", "The count is readable without zooming.")
		}, want: []string{"03:missing_criteria"}},
		{name: "unclosed fence", setup: func(r *repo) {
			r.edit(r.ticket(co, "03").Path, "## Verification", "```\n## Verification")
		}, want: []string{"03:malformed"}},
		{name: "title number mismatch", setup: func(r *repo) { r.edit(r.ticket(co, "03").Path, "# 03:", "# 07:") }, want: []string{"03:number_mismatch"}},
		{name: "duplicate ticket number", setup: func(r *repo) {
			r.write(".scratch/checkout/issues/03-copy.md", r.read(r.ticket(co, "03").Path))
		}, want: []string{"03:ambiguous_ticket"}},
		{name: "badly named file", setup: func(r *repo) { r.write(".scratch/checkout/issues/notes.md", "# Notes\n") }, want: []string{"bad_ticket_file"}},
		{name: "story gap", setup: func(r *repo) { r.edit(spec, "US 2:", "US 3:") }, want: []string{"spec:story_ids"}},
		{name: "missing history entry", setup: func(r *repo) { r.edit(spec, "Revision: 1", "Revision: 2") }, want: []string{"spec:missing_history"}},
		{name: "missing spec section", setup: func(r *repo) { r.edit(spec, "## Solution", "## Solutions") }, want: []string{"spec:missing_section"}},
		{name: "zero attempt limit", setup: func(r *repo) { r.edit(spec, "Attempt limit: 3", "Attempt limit: 0") }, want: []string{"spec:invalid_field"}},
		{name: "fenced spec text isn't metadata", setup: func(r *repo) {
			r.edit(spec, "- Add a count to the `summary` module.", "- From a prototype:\n\n```text\nRevision: 9\n## Fake\nUS 7: fake\n```")
		}},

		{name: "uncovered story", full: true, setup: func(r *repo) { r.setMeta(co, "02", "Covers", "US 1") }, want: []string{"story_not_covered"}},
		{name: "acceptance missing a story", full: true, setup: func(r *repo) { r.setMeta(co, "04", "Covers", "US 1") }, want: []string{"04:acceptance_coverage"}},
		{name: "acceptance of an old revision", full: true, setup: func(r *repo) {
			r.edit(spec, "Revision: 1", "Revision: 2")
			r.edit(spec, "- Revision 1: Initial spec.", "- Revision 1: Initial spec.\n- Revision 2: Reworded.")
		}, want: []string{"04:stale_acceptance"}},
		{name: "acceptance not blocked by every ticket", full: true, setup: func(r *repo) { r.setMeta(co, "04", "Blocked by", "01, 02") }, want: []string{"04:acceptance_blockers"}},
		{name: "no acceptance", full: true, setup: func(r *repo) { r.setMeta(co, "04", "Status", "cancelled") }, want: []string{"no_acceptance"}},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r := newRepo(t, co)
			c.setup(r)
			s := r.status(co)
			if s.Valid == (len(c.want) > 0) {
				t.Fatalf("valid=%v with errors %+v", s.Valid, s.Errors)
			}
			has := func(errs []problem, w string) bool {
				for _, e := range errs {
					if e.Code == w || e.Ticket+":"+e.Code == w {
						return true
					}
				}
				return false
			}
			for _, w := range c.want {
				if !has(s.Errors, w) {
					t.Fatalf("missing %s in %+v", w, s.Errors)
				}
			}
			for _, a := range c.absent {
				if has(s.Errors, a) {
					t.Fatalf("unexpected %s in %+v", a, s.Errors)
				}
			}

			res := r.run("", "validate", co)
			if (res.code == 0) != (len(c.want) == 0) {
				t.Fatalf("validate exit %d: %s", res.code, res.raw)
			}
			draft := r.run("", "validate", co, "--draft")
			var draftErrs []problem
			if draft.code != 0 {
				e := draft.err(t)
				for _, p := range e.Details["errors"].([]any) {
					m := p.(map[string]any)
					ticket, _ := m["ticket"].(string)
					draftErrs = append(draftErrs, problem{Code: m["code"].(string), Ticket: ticket})
				}
			}
			for _, w := range c.want {
				if has(draftErrs, w) == c.full {
					t.Fatalf("draft validation: %s present=%v, want %v", w, !c.full, !c.full)
				}
			}
		})
	}
}

func TestAmbiguousAndMissingDocuments(t *testing.T) {
	r := newRepo(t, co)
	r.write(".scratch/checkout/issues/01-copy.md", r.read(r.ticket(co, "01").Path))
	e := r.fails("ambiguous_ticket", "", "ticket", "get", co, "01")
	if len(e.Details["paths"].([]any)) != 2 {
		t.Fatalf("details: %+v", e.Details)
	}
	r.fails("ambiguous_ticket", "", "ticket", "claim", co, "01", "--if-match", "x", "--commit", shaA)
	r.fails("ticket_not_found", "", "ticket", "get", co, "09")
	r.fails("feature_not_found", "", "status", "missing")
	r.fails("feature_not_found", "", "spec", "get", "missing")
}

func TestUnsafePaths(t *testing.T) {
	r := newRepo(t, co)
	outside := t.TempDir()
	os.WriteFile(filepath.Join(outside, "spec.md"), []byte("# Outside\n"), 0o644)

	os.Symlink(outside, filepath.Join(r.root, ".scratch", "linked"))
	r.fails("unsafe_path", "", "spec", "get", "linked")
	r.fails("unsafe_path", `{"attempt_limit":4}`, "spec", "edit", "linked", "--if-match", "x", "--input", "-")

	os.MkdirAll(filepath.Join(r.root, ".scratch", "filelink"), 0o755)
	os.Symlink(filepath.Join(outside, "spec.md"), filepath.Join(r.root, ".scratch", "filelink", "spec.md"))
	r.fails("unsafe_path", "", "status", "filelink")
	if b, _ := os.ReadFile(filepath.Join(outside, "spec.md")); string(b) != "# Outside\n" {
		t.Fatal("wrote through a symlink")
	}

	for _, id := range []string{"../etc", ".hidden", "a/b", "a..b"} {
		r.fails("invalid_feature_id", "", "status", id)
	}
	var list struct {
		Features []struct{ Feature string } `json:"features"`
	}
	r.ok(&list, "", "feature", "list")
	if len(list.Features) != 1 || list.Features[0].Feature != co {
		t.Fatalf("feature list must skip symlinked features: %+v", list)
	}
}
