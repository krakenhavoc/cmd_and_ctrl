package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// A Killer Among Us — Enchantment {4}{G} (Murders at Karlov Manor,
// #4469):
//
//	"When this enchantment enters, create a 1/1 white Human creature
//	 token, a 1/1 blue Merfolk creature token, and a 1/1 red Goblin
//	 creature token. Then secretly choose Human, Merfolk, or Goblin.
//	 Sacrifice this enchantment, Reveal the creature type you chose: If
//	 target attacking creature token is the chosen type, put three +1/+1
//	 counters on it and it gains deathtouch until end of turn."
//
// The tokens are made, then the controller is asked, privately, to choose
// one of the three words (game.QueueSecretChoiceForEffect): it is stamped
// on the enchantment (Card.SecretChoice), carried by the snapshot, never
// shown in a view, and announced by nothing. The ability's cost reveals it
// (game.AbilityCost.RevealSecretChoice): the word is shown to the table as
// the cost is paid, so it is public even if the ability is then countered
// or its target goes illegal, and only the player who chose may pay it (a
// stolen enchantment can't be revealed by its new controller). The
// enchantment is sacrificed in the same cost, so the effect reads the
// choice as the enchantment last existed (CR 608.2h). It can target ANY
// attacking creature token, not only its own three; only one of the chosen
// type gets the counters and deathtouch. Two copies each keep their own
// choice.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "79d577a8-53c0-4220-bc0e-70764d72d5fd",
		Name:         "A Killer Among Us",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("A Killer Among Us — create a Human, a Merfolk and a Goblin, then secretly choose one of the types",
				Do(
					CreateToken{Template: TokenCard("1/1 white Human"), N: 1},
					CreateToken{Template: TokenCard("1/1 blue Merfolk"), N: 1},
					CreateToken{Template: RedGoblinToken(), N: 1},
					Instruction(func(ctx *Context) error {
						ctx.Game.QueueSecretChoiceForEffect(ctx.Controller(), ctx.Source(),
							"A Killer Among Us — secretly choose Human, Merfolk, or Goblin (only you see the answer until you reveal it)",
							[]string{"Human", "Merfolk", "Goblin"})
						return nil
					}))),
		},
		Activated: []ActivatedAbility{{
			Label: "Sacrifice this enchantment, Reveal the creature type you chose: If target attacking creature token is the chosen type, " +
				"put three +1/+1 counters on it and it gains deathtouch until end of turn.",
			Purpose: game.Purpose{
				Answers: game.AnswerPump | game.AnswerCombatGrant,
				Targets: game.ForTargets(game.TargetPurpose{
					Slot: 0, UntilEndOfTurn: game.TargetBenefit,
					Pump: &game.Pump{Power: 3, Toughness: 3, Keywords: []string{"deathtouch"}},
				}),
			},
			Cost:      Plus(SacrificeThis(), RevealSecretChoice()),
			Condition: YouSecretlyChoseForThis,
			Targets:   TargetCreature("target attacking creature token", AttackingCreature(), IsTokenPredicate()),
			Effect:    aKillerAmongUsStrike,
		}},
	})
}

// aKillerAmongUsStrike is the resolution: the chosen type is read from the
// enchantment as it last existed, and the target, if it is still a legal
// attacking token of that type, gets three +1/+1 counters and deathtouch.
func aKillerAmongUsStrike(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	src, ok := ctx.SourcePermanent()
	if !ok || src.SecretChoice == "" {
		return nil
	}
	ts := ctx.LegalTargets()
	if len(ts) == 0 {
		return nil
	}
	target, ok := g.LookupCardForEffect(ts[0].ID)
	if !ok || !target.HasSubtype(src.SecretChoice) {
		return nil
	}
	if err := (AddCounter{Target: ts[0].ID, Kind: game.CounterPlusOne, N: 3}).Apply(ctx); err != nil {
		return err
	}
	return GrantKeywordUntilEOT{Target: ts[0].ID, Keywords: []string{"deathtouch"}, Label: "A Killer Among Us — deathtouch"}.Apply(ctx)
}
