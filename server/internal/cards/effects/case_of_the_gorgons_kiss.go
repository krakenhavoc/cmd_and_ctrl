package effects

import (
	"slices"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Case of the Gorgon's Kiss — Enchantment — Case {B} (Murders at Karlov
// Manor, #4469):
//
//	"When this Case enters, destroy up to one target creature that was dealt
//	 damage this turn.
//	 To solve — Three or more creature cards were put into graveyards from
//	 anywhere this turn. (If unsolved, solve at the beginning of your end
//	 step.)
//	 Solved — This Case is a 4/4 Gorgon creature with deathtouch and
//	 lifelink in addition to its other types."
//
// The target is a creature that was dealt damage as of resolution (the
// permanent that entered after the damage is a new object that was dealt
// none, CR 400.7), "up to one" so it may have none. To solve counts the
// creature cards put into any graveyard from anywhere this turn, tokens (not
// cards) never, each judged by the type the card has in the graveyard (the
// ruling), at the end step and again on resolution (CR 603.4). Solved is three
// statics gated on the Case's own solved designation (ActiveWhen: Solved()):
// the Creature type and Gorgon subtype in addition to Enchantment and Case
// (layer 4), a 4/4 base power and toughness (layer 7b), deathtouch and
// lifelink (layer 6). It stays solved for as long as it is on the battlefield
// (CR 719.3b).
//
// No simplification.
func init() {
	isSelf := func(target *game.Card, _ *game.Game, source *game.Card) bool {
		return target.InstanceID == source.InstanceID
	}
	keyword := func(kw string) game.StaticAbility {
		ab := KeywordGrant(isSelf, kw)
		ab.Label = "Has " + kw + " while solved."
		ab.ActiveWhen = Solved()
		return ab
	}
	Register(Spec{
		OracleID:     "a89f9072-cf06-429f-a9b8-c11b5731ddd3",
		Name:         "Case of the Gorgon's Kiss",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Case of the Gorgon's Kiss — destroy up to one target creature that was dealt damage this turn",
				destroyChosenPermanent),
				TargetCreature("up to one target creature that was dealt damage this turn", DealtDamageThisTurn()).WithCount(0, 1)),
			ToSolve("Case of the Gorgon's Kiss — to solve: three or more creature cards were put into graveyards from anywhere this turn",
				threeOrMoreCreatureCardsWentToGraveyardsThisTurn),
		},
		Static: []game.StaticAbility{
			{
				Label:      "This Case is a Gorgon creature while solved.",
				Layer:      game.Layer4Type,
				AppliesTo:  isSelf,
				ActiveWhen: Solved(),
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, source *game.Card) {
					if !slices.Contains(c.Types, "Creature") {
						c.Types = append(c.Types, "Creature")
					}
					if gorgon := source.TextSubtype("Gorgon"); !slices.Contains(c.Subtypes, gorgon) {
						c.Subtypes = append(c.Subtypes, gorgon)
					}
				},
			},
			{
				Label:      "This Case is a 4/4 while solved.",
				Layer:      game.Layer7PT,
				SubLayer:   game.SubLayer7B_Set,
				AppliesTo:  isSelf,
				ActiveWhen: Solved(),
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					c.Power, c.Toughness = 4, 4
				},
			},
			keyword("deathtouch"),
			keyword("lifelink"),
		},
	})
}

// threeOrMoreCreatureCardsWentToGraveyardsThisTurn is the solve condition:
// at least three zone moves into a graveyard this turn of cards that are
// creature cards where they are now. A token is not a card and ceases to
// exist, so it is never found and never counted.
func threeOrMoreCreatureCardsWentToGraveyardsThisTurn(g *game.Game, _, _ uuid.UUID) bool {
	n := 0
	for _, ev := range g.EventsThisTurn() {
		if ev.Kind != game.EventZoneMove || ev.NewZone != game.ZoneGraveyard || ev.CardID == uuid.Nil {
			continue
		}
		if c, ok := g.LookupCardForEffect(ev.CardID); ok && !c.IsToken() && c.IsCreature() {
			n++
		}
	}
	return n >= 3
}
