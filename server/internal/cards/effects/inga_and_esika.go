package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Inga and Esika — Legendary Creature — Human God {2}{G}{U}, 4/4:
//
//	"Creatures you control have vigilance and "{T}: Add one mana of any
//	 color. Spend this mana only to cast a creature spell."
//	 Whenever you cast a creature spell, if three or more mana from
//	 creatures was spent to cast it, draw a card."
//
// Two layer 6 grants over every creature its controller controls, Inga and
// Esika included: vigilance, and a granted mana ability (ADR 0093) whose
// mana carries the spend rider "only to cast a creature spell" (the
// Beastcaller Savant / Clement shape), so it can pay for nothing else.
//
// The cast trigger reads the spell's payment record, where each mana keeps
// a snapshot of what its source was when it was minted (#1212,
// ManaSourceCreature): a creature that was sacrificed or died after
// tapping still counts, and a land the engine animated into a creature
// counts too. The condition is on the trigger (CR 603.4): the payment is a
// fact about the cast, so it cannot change before the draw. Mana paid
// outside the pool is asked for per mana (ReadsSourceKinds), as Coin of
// Mastery's is.
//
// No simplification.
const ingaAndEsikaGrant = "inga-and-esika/creature-only-any-color"

func init() {
	Register(Spec{
		OracleID:       "3431ee42-c34a-4972-9488-41b83ba5f943",
		Name:           "Inga and Esika",
		Completeness:   CompletenessFull,
		ReadsManaSpent: game.ReadsYourCasts | game.ReadsSourceKinds,
		Grants: []AbilityGrant{{
			Key: ingaAndEsikaGrant,
			Mana: []ManaAbility{{
				Cost:         ManaAbilityCost{Tap: true},
				Produced:     "{W|U|B|R|G}",
				Restrictions: []string{ManaRestrictCast, ManaRestrictType("Creature")},
				Label:        "Add one mana of any color. Spend this mana only to cast a creature spell",
			}},
			Text: "{T}: Add one mana of any color. Spend this mana only to cast a creature spell.",
		}},
		Static: []game.StaticAbility{
			Labelled(KeywordGrant(b16CreaturesYouControl, "vigilance"),
				"Creatures you control have vigilance."),
			Labelled(GrantAbilities(b16CreaturesYouControl, ingaAndEsikaGrant),
				"Creatures you control have \"{T}: Add one mana of any color. Spend this mana only to cast a creature spell.\""),
		},
		Triggered: []game.TriggeredAbility{
			func() game.TriggeredAbility {
				cast := YouCast(Creature())
				return On(game.EventCast, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
					return cast(ev, source, lki, g) &&
						g.StackItemPaidForEffect(ev.CardID).Spent().CountFrom(game.ManaSourceCreature) >= 3
				}, "Inga and Esika — draw a card", Do(DrawCards{N: 1}))
			}(),
		},
	})
}
