package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// closure_doors_token_restore_test.go — #4253 slice 6, the token half: the
// "then" of a token creation is a registered tail, so the CR 616 order
// prompt a Doubling Season beside a Parallel Lives asks is a restore point.
// Each test takes a STRICT restore while the prompt is open, answers it on
// the restored table and checks the card still finishes off the disk.
// (The helpers libTop and tokenOrderBoard are in
// closure_doors_restore_test.go.)

// orderPromptOpen reports whether the one open prompt is a CR 616 order.
func OrderPromptOpen(g *game.Game) bool {
	return len(g.PendingChoices) == 1 && g.PendingChoices[0].Kind == game.PendingChoiceReplacementOrder
}

// tokenEntryOrderProbeOracles are two test cards, each "if a token would
// enter the battlefield, it enters instead": two different declared entry
// replacements, so a token's ENTRY asks for their order (CR 616.1). They
// are catalog rows, so a restored table gathers them again. #4994: the
// token a copy of a permanent spell becomes is not created (CR 111.13), so
// its CR 616 pause can no longer come from a Doubling Season beside a
// Parallel Lives; it comes from its entry, which these give it.
var tokenEntryOrderProbeOracles = []string{"test-4994-token-entry-order-a", "test-4994-token-entry-order-b"}

func init() {
	for i, oracle := range tokenEntryOrderProbeOracles {
		effects.Register(effects.Spec{
			OracleID: oracle,
			Name:     "Token Entry Order Probe " + string(rune('A'+i)),
			Replacements: []game.ReplacementEffect{{
				Watches: []game.EventKind{game.EventZoneMove},
				AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, _ *game.Card) bool {
					if ev.Kind != game.RepEventMove || ev.NewZone != game.ZoneBattlefield {
						return false
					}
					c, ok := g.LookupCardForEffect(ev.CardID)
					return ok && c.IsToken()
				},
				Replace: func(*game.ReplacementEvent, *game.Game, *game.Card) error { return nil },
				Label:   "Token Entry Order Probe " + string(rune('A'+i)),
			}},
		})
	}
}

// seedTokenEntryOrderProbes puts both probes on the battlefield under
// `who` and returns them in order.
func SeedTokenEntryOrderProbes(g *game.Game, who uuid.UUID) (a, b uuid.UUID) {
	a = SeedReplacementPermanent(g, tokenEntryOrderProbeOracles[0], "Token Entry Order Probe A", who)
	b = SeedReplacementPermanent(g, tokenEntryOrderProbeOracles[1], "Token Entry Order Probe B", who)
	return a, b
}

// tokenCreatedEventsFor counts the EventTokenCreated events naming id.
func TokenCreatedEventsFor(g *game.Game, id uuid.UUID) int {
	n := 0
	for _, ev := range g.Events {
		if ev.Kind == game.EventTokenCreated && ev.CardID == id {
			n++
		}
	}
	return n
}
