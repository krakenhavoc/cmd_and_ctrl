package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Concerted Effort — Enchantment {2}{W}{W}:
//
//	"At the beginning of each upkeep, creatures you control gain flying
//	 until end of turn if a creature you control has flying. The same is
//	 true for fear, first strike, double strike, landwalk, protection,
//	 trample, and vigilance."
//
// One triggered ability, firing at the start of every player's upkeep.
// As it resolves it reads the keywords the creatures its controller
// controls have right now, from their effective abilities (so a granted
// or lost keyword counts), and gives each of the ones found to every
// creature that controller controls until end of turn. The set of
// creatures is locked as the ability resolves (CR 611.2c), and every
// keyword is read before any is granted, so the keywords just granted
// are not themselves shared again.
//
// Landwalk and protection are families, not single keywords. Each
// distinct landwalk a creature has (swampwalk, nonbasic landwalk, ...)
// is shared as that landwalk, and each distinct protection quality as
// protection from that quality: a creature with protection from red and
// another with protection from artifacts give every creature both. A
// quality that names a player ("protection from the chosen player") is
// shared as protection from that same player (CR 702.16k).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fd4236e5-6a2c-45ec-afb3-cd4fc516fbf1",
		Name:         "Concerted Effort",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtEachStep(game.StepUpkeep,
				"Concerted Effort — creatures you control gain each keyword a creature you control has, until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					keywords := concertedEffortKeywords(g, item.Controller)
					if len(keywords) == 0 {
						return nil
					}
					controller := item.Controller
					return GrantKeywordUntilEOT{
						Keywords: keywords,
						Label:    "Concerted Effort — shared keywords",
						Match: func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
							return c.IsCreature() && c.Controller == controller
						},
					}.Apply(ctx)
				}),
		},
	})
}

// concertedEffortShared is the single-word keywords the card names.
var concertedEffortShared = []string{"flying", "fear", "first strike", "double strike", "trample", "vigilance"}

// concertedEffortKeywords is the keyword tokens, without repeats, that
// at least one creature `controller` controls has: the plain ones named
// on the card, each landwalk, and each protection quality.
func concertedEffortKeywords(g *game.Game, controller uuid.UUID) []string {
	seen := map[string]bool{}
	var out []string
	add := func(token string) {
		if token != "" && !seen[token] {
			seen[token] = true
			out = append(out, token)
		}
	}
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller != controller || !c.IsCreature() {
			continue
		}
		card := c
		for _, kw := range concertedEffortShared {
			if game.HasKeyword(&card, kw) {
				add(kw)
			}
		}
		for _, a := range card.Effective().Abilities {
			if game.IsLandwalkToken(a) {
				add(a)
			}
		}
		for _, q := range game.ProtectionQualities(&card) {
			if q.Kind == game.ProtectionQualityPlayer {
				// The grantee has made no choice of its own: name the seat.
				if q.Player != uuid.Nil {
					add(game.ProtectionFromPlayer(q.Player))
				}
				continue
			}
			add(q.Token())
		}
	}
	return out
}
