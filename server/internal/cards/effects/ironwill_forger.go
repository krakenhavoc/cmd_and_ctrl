package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ironwill Forger — Creature — Orc Artificer {3}{W}, 3/3 (Tarkir:
// Dragonstorm Commander):
//
//	"Lieutenant — At the beginning of combat on your turn, if you control
//	 your commander, target nonlegendary creature you control gains myriad
//	 until end of turn. (Whenever it attacks, for each opponent other than
//	 defending player, you may create a token copy that's tapped and attacking
//	 that player or a planeswalker they control. Exile the tokens at end of
//	 combat.)"
//
// "If you control your commander" is an intervening if (CR 603.4), the
// lieutenant read (b18ControlsYourCommander: a commander you OWN and control),
// checked as the step begins and again as the ability resolves. The grant is
// myriad's shared bundle (GrantedMyriad, Muddle's) given to the target until
// end of turn; the target is judged again on resolution (CR 608.2b), so a
// creature that became legendary or left is not given myriad.
//
// No simplification.
func init() {
	label := "Ironwill Forger — target nonlegendary creature you control gains myriad until end of turn"
	Register(Spec{
		OracleID:     "eee99fdc-b116-4f06-9f9f-7b7dda5743c1",
		Name:         "Ironwill Forger",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key:       ironwillForgerMyriadGrant,
			Text:      "Myriad",
			Triggered: []game.TriggeredAbility{GrantedMyriad("Ironwill Forger")},
		}},
		Triggered: []game.TriggeredAbility{
			Targeting(On(game.EventStepBegan, AllOf(StepBegan(game.StepBeginCombat, true),
				func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return b18ControlsYourCommander(g, source.Controller)
				}), label, func(g *game.Game, item *game.StackItem) error {
				if !b18ControlsYourCommander(g, item.Controller) {
					return nil
				}
				ctx := NewContext(g, item)
				target, ok := ctx.ClauseTarget(0)
				if !ok {
					return nil
				}
				return GrantAbilitiesFor{Target: target.ID, Keys: []string{ironwillForgerMyriadGrant},
					Label: "Ironwill Forger — myriad until end of turn"}.Apply(ctx)
			}),
				PermanentYouControl("target nonlegendary creature you control", Creature(), Not(Legendary()))),
		},
	})
}

const ironwillForgerMyriadGrant = "ironwill-forger/myriad"
