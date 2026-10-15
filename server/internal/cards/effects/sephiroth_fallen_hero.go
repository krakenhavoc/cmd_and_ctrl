package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sephiroth, Fallen Hero — Legendary Creature — Human Avatar Soldier
// {3}{R}{W}, 7/5 (#3656):
//
//	"Jenova Cells — Whenever Sephiroth attacks, you may put a cell
//	 counter on target creature. Until end of turn, each modified
//	 creature you control has base power and toughness 7/5. (Equipment,
//	 Auras you control, and counters are modifications.)
//	 The Reunion — {3}, Sacrifice a modified creature: Return this card
//	 from your graveyard to the battlefield tapped."
//
// The attack trigger is mandatory and targets; the "may" is the counter
// alone: the question is asked as the trigger resolves, and the base-P/T
// sentence happens either way. The set of modified creatures is fixed as the effect applies
// (CR 611.2c), after the counter, so the creature that just got the
// counter is in it. The base-P/T set sits in layer 7b, so counters and
// Equipment still add to the 7/5. The Reunion is Reassembling Skeleton's
// return from the graveyard with a sacrifice cost, paid at announce.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ca9ab230-51e2-4ef2-a7b9-4e325d9f9657",
		Name:         "Sephiroth, Fallen Hero",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				WheneverThisAttacks("Sephiroth, Fallen Hero — Jenova Cells: you may put a cell counter on target creature; each modified creature you control has base power and toughness 7/5 until end of turn",
					sephirothJenovaCells),
				TargetCreature("target creature")),
		},
		Activated: []ActivatedAbility{{
			Label:   "{3}, Sacrifice a modified creature: Return this card from your graveyard to the battlefield tapped.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet | game.AnswerProtect},
			Cost: Plus(ManaCost("{3}"), game.AbilityCost{
				SacrificeOther: sacrificeSpec("a modified creature", Creature(), YouControl(), modifiedCreature),
			}),
			Zones:  []game.ZoneKind{game.ZoneGraveyard},
			Effect: returnThisFromGraveyardTapped,
		}},
	})
}

func sephirothJenovaCells(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	target, hasTarget := b16FirstLegalTargetCard(ctx)
	base := func(g *game.Game) error {
		bctx := NewContext(g, item)
		return ScopedEffectFor{
			Match: And(Creature(), YouControl(), modifiedCreature),
			Mods: []game.Mod{
				game.SetBasePowerMod(7),
				game.SetBaseToughnessMod(5),
			},
			Duration: DurationUntilEndOfTurn(bctx),
			Label:    "Sephiroth, Fallen Hero — each modified creature you control has base power and toughness 7/5 until end of turn",
		}.Apply(bctx)
	}
	if !hasTarget {
		return base(g)
	}
	g.QueueConfirmForEffect(game.ConfirmPrompt{
		Chooser:  item.Controller,
		Source:   item.SourceCardID,
		Question: "Sephiroth, Fallen Hero — put a cell counter on target creature?",
		OnAccept: func(g *game.Game) error {
			if err := (AddCounter{Target: target, Kind: "cell", N: 1}).Apply(NewContext(g, item)); err != nil {
				return err
			}
			return base(g)
		},
		OnDecline: base,
	})
	return nil
}
