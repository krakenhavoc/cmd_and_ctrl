package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gimli's Reckless Might — Enchantment {3}{R} (EDHREC rank 4465):
//
//	"Creatures you control have haste.
//	 Formidable — Whenever you attack, if creatures you control have
//	 total power 8 or greater, target attacking creature you control
//	 fights up to one target creature you don't control."
//
// A four-mana anthem that makes a go-wide board immediate and then
// eats a blocker every combat. Haste on its own is why a Gruul deck
// runs it — every creature it casts attacks the turn it lands — and
// the fight is the payoff for the board that haste built.
//
// TWO TARGET CLAUSES with different predicates, which the engine
// could not express until #937 (ADR 0065): slot 0 is an ATTACKING
// creature YOU control, slot 1 is "up to one" creature you DON'T
// control. Each is checked against its own predicate at announce
// (CR 601.2c) and re-checked against its own at resolution
// (CR 608.2b), so a fighter removed in response does nothing and a
// victim removed in response leaves the fighter unharmed. The second
// clause really is "up to one": declining it is legal and the fight
// simply does not happen.
//
// FORMIDABLE IS AN INTERVENING-IF (CR 603.4): total power is checked
// when the trigger would go on the stack and again as it resolves, so
// a creature killed after blockers really does turn the fight off.
// Total power is CurrentPower summed over the creatures the
// controller has — counters, anthems and pumps all count (CR 208.3) —
// and it counts EVERY creature, not only the attackers, which is what
// "creatures you control" means.
//
// "Whenever you attack" is ONE trigger per attack declaration, not one
// per attacker: the engine emits an EventAttack per creature, so
// OncePerBatch declines every later event of the same batch. Without
// it a four-creature alpha strike would fight four times, which is the
// #259 direction.
//
// The fight is b10Fight, which reads both creatures' power before
// either amount lands, so two lethal fighters trade rather than the
// first one killing the second and surviving.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "824abca7-b5e2-4ddb-ac7d-b7d04758a3c9",
		Name:         "Gimli's Reckless Might",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			b16GrantKeywords(b16CreaturesYouControl, "haste"),
		},
		Triggered: []game.TriggeredAbility{
			OncePerBatch(game.TriggeredAbility{
				Watches: []game.EventKind{game.EventAttack},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return attackDeclaredByYou(ev, source.Controller) &&
						b43TotalPowerControlled(g, source.Controller) >= 8
				},
				Targets: Clauses(
					TargetCreature("target attacking creature you control",
						And(YouControl(), AttackingCreature())),
					Distinct(TargetCreature("up to one target creature you don't control",
						OpponentControls()).WithCount(0, 1)),
				),
				Key: "Gimli's Reckless Might — the attacker fights a creature you don't control",
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					// CR 603.4: the intervening if is re-checked as
					// the ability resolves.
					if b43TotalPowerControlled(g, item.Controller) < 8 {
						return nil
					}
					mine, ok := ctx.ClauseTarget(0)
					if !ok || mine.Kind != game.TargetCard {
						return nil
					}
					theirs, ok := ctx.ClauseTarget(1)
					if !ok || theirs.Kind != game.TargetCard {
						// "Up to one" — declined, or gone. The
						// fight does not happen and nothing else
						// on the card does either.
						return nil
					}
					return b10Fight(ctx, mine.ID, theirs.ID)
				},
			}),
		},
	})
}
