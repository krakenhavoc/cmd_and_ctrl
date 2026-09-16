package legal_test

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// search_validate_test.go pins the enumerator's one promise where it
// used to break: a search that carries a SearchLibrarySpec.Validate
// hook.
//
// The promise is the package doc's — every Move EnumerateFor returns
// would be ACCEPTED by actions.Dispatch. A Validate hook expresses a
// clause no per-card predicate can ("two basic lands that SHARE a
// land type"), it lives on the unexported continuation frame, and
// the enumerator used to have no way to ask about it, so it offered
// pairs the resolver refuses. The resolver refuses without dequeuing
// — correct, so a human can retry — and a deterministic bot retried
// identically until the table stopped (#544).

const oracleMyriadLandscape = "2549bc57-9ffb-4053-9f10-f2a5f792b845"

// openMyriadSearch stands a Myriad Landscape up on the active seat's
// battlefield with the mana to crack it, stocks the library with
// `forests` Forests and `islands` Islands, activates the fetch
// through the enumerator and resolves it, leaving the search prompt
// open.
//
// The two basic TYPES are the point: the fetch's own predicate
// accepts every one of them, so every single-card pick is legal and
// exactly the mixed PAIRS are not. They are interleaved so the first
// two candidates in library order are a mismatched pair, which is the
// shape the field replay wedged on — the lexicographically first pair
// is the one the old enumerator had budget for, and it is invalid.
func openMyriadSearch(t *testing.T, forests, islands int) (*game.Game, *game.Player) {
	t.Helper()
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	land := battlefieldCard(g, active, game.Card{
		Name:     "Myriad Landscape",
		TypeLine: "Land",
		OracleID: oracleMyriadLandscape,
	})
	// Two Mountains pay the {2}. A third basic type, so they can
	// never be confused with the library's candidates.
	for i := 0; i < 2; i++ {
		battlefieldCard(g, active, basic("Mountain", "Mountain"))
	}
	push := func(name, sub string) {
		c := basic(name, sub)
		c.InstanceID, c.Owner, c.Controller = uuid.New(), active.ID, active.ID
		active.Library.PushTop(c)
	}
	for i := 0; i < forests || i < islands; i++ {
		if i < forests {
			push("Forest", "Forest")
		}
		if i < islands {
			push("Island", "Island")
		}
	}
	advanceTo(t, g, game.StepPrecombatMain)

	moves := legal.EnumerateFor(g, active.ID)
	var crack *legal.Move
	for i := range moves {
		if moves[i].Source == land && moves[i].Kind == legal.KindActivate {
			crack = &moves[i]
			break
		}
	}
	if crack == nil {
		t.Fatalf("Myriad Landscape's fetch was not offered: %v", labels(moves))
	}
	if err := actions.Dispatch(g, actions.Action{
		Type: actions.Type(crack.Type), Player: active.ID, Caller: active.ID, Params: crack.Params,
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	for i := 0; i < 8 && len(g.PendingChoices) == 0; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("pass: %v", err)
		}
	}
	if len(g.PendingChoices) != 1 || g.PendingChoices[0].Kind != game.PendingChoiceSearchLibrary {
		t.Fatalf("expected a search prompt, got %+v", g.PendingChoices)
	}
	return g, active
}

// pickedCardIDs reads a resolve_choice move's card_ids back off the
// wire params rather than off the label, because the params are what
// the dispatcher will act on.
func pickedCardIDs(t *testing.T, m legal.Move) []uuid.UUID {
	t.Helper()
	var p struct {
		CardIDs []string `json:"card_ids"`
	}
	if err := json.Unmarshal(m.Params, &p); err != nil {
		t.Fatalf("params of %q: %v", m.Label, err)
	}
	out := make([]uuid.UUID, 0, len(p.CardIDs))
	for _, s := range p.CardIDs {
		id, err := uuid.Parse(s)
		if err != nil {
			t.Fatalf("card id %q in %q: %v", s, m.Label, err)
		}
		out = append(out, id)
	}
	return out
}

func libraryNames(p *game.Player) map[uuid.UUID]string {
	out := make(map[uuid.UUID]string, len(p.Library.Cards))
	for _, c := range p.Library.Cards {
		out[c.InstanceID] = c.Name
	}
	return out
}

// TestSearchEnumerationNeverOffersAPickTheEngineRejects is the
// general invariant, stated once and loudly: for a search carrying a
// Validate hook, every enumerated answer is one the dispatcher
// accepts.
//
// dispatchAll is the assertion — it sends each move against a fresh
// clone — and the mixed-pair check below says out loud which
// rejection this scenario is about, so a regression reads as "the
// enumerator offered Forest+Island" rather than as an opaque
// ErrInvalidParam.
func TestSearchEnumerationNeverOffersAPickTheEngineRejects(t *testing.T) {
	g, active := openMyriadSearch(t, 4, 4)
	moves := legal.EnumerateFor(g, active.ID)
	names := libraryNames(active)

	// The invariant.
	dispatchAll(t, g, active.ID, moves)

	pairs := 0
	for _, m := range moves {
		if m.Kind != legal.KindChoice {
			continue
		}
		ids := pickedCardIDs(t, m)
		if len(ids) < 2 {
			continue
		}
		pairs++
		first := names[ids[0]]
		for _, id := range ids[1:] {
			if names[id] != first {
				t.Errorf("offered a pick that breaks the share-a-land-type clause: %q", m.Label)
			}
		}
	}
	// Without at least one pair on offer this test proves nothing —
	// the hook is only consulted for picks of two or more.
	if pairs == 0 {
		t.Fatalf("no two-card pick was offered, so the Validate hook was never exercised: %v", labels(moves))
	}
}

// TestSearchForUpToTwoReachesValidPairsAtEveryLibrarySize is the
// play-quality half of #544, and the half the field evidence
// sharpened.
//
// The old combinations() walked pick sizes in order and spent the
// whole MaxExpansionPerSource budget on one-card picks before it
// reached the first two-card pick. The boundary was hard and the
// failure flipped across it:
//
//   - 12 or more candidates: the budget was gone before k=2, so NO
//     pair was offered and "search for up to two basic lands" quietly
//     became "search for one" — for bots and for any client driving
//     off this list.
//   - 11 or fewer: exactly one or two pairs squeezed in, always the
//     first in library order, and roughly half the time that pair
//     breaks the share-a-land-type clause. With 11 there was no
//     second pair to fall back on, which is how the live table died.
//
// So filtering the invalid pair out is necessary and NOT sufficient:
// it would leave a legal list that still never offers a valid pair.
// The budget has to REACH the valid pairs, which is why it is now
// spread across pick sizes rather than spent smallest-first.
//
// The sizes below bracket the cap, and include the field shape (the
// curated simic-ramp deck's 7 Forest / 5 Island in 99 cards, one
// Forest played).
func TestSearchForUpToTwoReachesValidPairsAtEveryLibrarySize(t *testing.T) {
	for _, tc := range []struct {
		name             string
		forests, islands int
	}{
		{"field-shape_6F_5I", 6, 5},
		{"simic-ramp_7F_5I", 7, 5},
		{"at-the-cap_6F_6I", 6, 6},
		{"over-the-cap_20F_20I", 20, 20},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g, active := openMyriadSearch(t, tc.forests, tc.islands)
			names := libraryNames(active)
			moves := legal.EnumerateFor(g, active.ID)

			// Nothing offered may be refused, at any size.
			dispatchAll(t, g, active.ID, moves)

			singles, pairs := 0, 0
			for _, m := range moves {
				if m.Kind != legal.KindChoice {
					continue
				}
				ids := pickedCardIDs(t, m)
				switch len(ids) {
				case 1:
					singles++
				case 2:
					pairs++
					if names[ids[0]] != names[ids[1]] {
						t.Errorf("offered a pick that breaks the share-a-land-type clause: %q", m.Label)
					}
				}
			}
			t.Logf("candidates=%d answers=%d singles=%d pairs=%d",
				len(g.PendingChoices[0].SearchCards), len(moves), singles, pairs)
			if pairs == 0 {
				t.Errorf("no valid two-card pick was reached: %d singles, 0 pairs", singles)
			}
			if singles == 0 {
				t.Errorf("taking a single land is still a legal answer and must stay on offer")
			}
		})
	}
}

// TestSearchOffersAnAlwaysLegalAnswer pins the escape hatch's
// enumerator half: a search prompt declares exactly one answer the
// engine cannot refuse, and it is "fail to find" (CR 701.23b).
// aiseat.SafeIndex is what reads this, and a seat owing a choice is
// offered nothing else — no pass — so without it a bot whose every
// other answer bounces has nowhere to go.
func TestSearchOffersAnAlwaysLegalAnswer(t *testing.T) {
	g, active := openMyriadSearch(t, 4, 4)
	moves := legal.EnumerateFor(g, active.ID)

	safe := -1
	marked := 0
	for i, m := range moves {
		if m.AlwaysLegal {
			marked++
			if safe < 0 {
				safe = i
			}
		}
	}
	if marked != 1 {
		t.Fatalf("want exactly one always-legal answer, got %d: %v", marked, labels(moves))
	}
	if len(pickedCardIDs(t, moves[safe])) != 0 {
		t.Errorf("the always-legal answer must be the empty pick, got %q", moves[safe].Label)
	}
	// It has to be answerable on its own, not only as part of the
	// batch dispatchAll sends.
	clone := g.Clone()
	if err := actions.Dispatch(clone, actions.Action{
		Type:   actions.Type(moves[safe].Type),
		Player: active.ID,
		Caller: active.ID,
		Params: moves[safe].Params,
	}); err != nil {
		t.Fatalf("the always-legal answer was refused: %v", err)
	}
	if len(clone.PendingChoices) != 0 {
		t.Errorf("failing to find must clear the prompt, %d still open", len(clone.PendingChoices))
	}
}

// TestPassPriorityIsMarkedAlwaysLegal: passing is the other answer
// nothing can refuse, and PassIndex and SafeIndex have to agree about
// it or the runner's rescue takes two different exits on the same
// board.
func TestPassPriorityIsMarkedAlwaysLegal(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	advanceTo(t, g, game.StepPrecombatMain)
	for _, m := range legal.EnumerateFor(g, active.ID) {
		if m.Kind == legal.KindPass && !m.AlwaysLegal {
			t.Fatalf("pass_priority is not marked always-legal")
		}
	}
}
