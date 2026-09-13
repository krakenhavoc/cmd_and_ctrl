package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Adeline, Resplendent Cathar — Legendary Creature — Human Knight
// {1}{W}{W}, */4 (EDHREC rank 493):
//
//	"Vigilance
//	 Adeline's power is equal to the number of creatures you control.
//	 Whenever you attack, for each opponent, create a 1/1 white Human
//	 creature token that's tapped and attacking that player or a
//	 planeswalker they control."
//
// Three abilities, all real:
//
//   - Vigilance rides PrintedKeywords.
//   - The power is a Layer 7a characteristic-defining ability
//     (Tarmogoyf's shape): self-only, SETS power to the creature
//     count on every recompute, which includes Adeline herself and
//     every token she just made. The printed "*" is 0 in the card
//     data; the CDA overrides it.
//   - "Whenever you attack" is ONE trigger per attack declaration.
//     The engine emits EventAttack per creature, so the AppliesTo
//     declines any further event while an Adeline trigger is already
//     queued or on the stack — b04TriggerPendingOrOnStack, a wider
//     net than Professional Face-Breaker's because attackers
//     declared one at a time each drain the queue (see the helper) —
//     and without it a three-creature attack would make three times
//     the tokens, which is the #259 direction.
//
// The tokens really do enter TAPPED AND ATTACKING: the template
// carries Tapped and AttackingTarget, CreateTokenForEffect copies
// both, and the combat damage step reads AttackingTarget off the
// battlefield, so each token deals its 1 to its player this combat
// and can be blocked in the declare-blockers step. They were never
// DECLARED as attackers, so "whenever a creature attacks" triggers
// (Hellrider) do not fire for them — CR 508.4, as printed. The
// opponents are read when the trigger resolves; a player eliminated
// in response gets no token.
//
// Sandbox simplification: "that player OR A PLANESWALKER THEY
// CONTROL" is always the player — the engine has no
// attack-a-planeswalker path (Hellrider's note), so the choice
// cannot arise yet.
func init() {
	Register(Spec{
		OracleID:        "38515f89-348b-4cf3-b7bd-1f6fe4ce2fba",
		Name:            "Adeline, Resplendent Cathar",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Static: []game.StaticAbility{{
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7A_CDA,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID
			},
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				c.Power = b04CreaturesControlled(g, source.Controller)
			},
		}},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventAttack},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return attackDeclaredByYou(ev, source.Controller) && !b04TriggerPendingOrOnStack(g, source)
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Adeline — a tapped and attacking Human for each opponent",
					func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						for _, opp := range ctx.Opponents() {
							tmpl := WhiteHumanToken()
							tmpl.Tapped = true
							tmpl.AttackingTarget = opp
							if err := (CreateToken{Controller: item.Controller, Template: tmpl, N: 1}).Apply(ctx); err != nil {
								return err
							}
						}
						return nil
					})
			},
		}},
	})
}
