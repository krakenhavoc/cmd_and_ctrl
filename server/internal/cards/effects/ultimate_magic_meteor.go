package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ultimate Magic: Meteor — Sorcery {5}{R} (#3656):
//
//	"Ultimate Magic: Meteor deals 7 damage to each creature. If this
//	 spell was cast from exile, for each opponent, choose an artifact
//	 or land that player controls. Destroy the chosen permanents.
//	 Foretell {5}{R}"
//
// Pyroclasm's damage sweep at 7 and foretell as one special action. The
// exile clause reads the spell's cast zone (CastFromZone), which a
// foretold cast and any other cast from exile both satisfy, as printed.
// The caster picks one artifact or land per opponent in seat order (an
// opponent with none is skipped), and all the chosen permanents are
// destroyed together once the last pick is in.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "31cc5d9d-9384-4a9b-ba7f-6a4edbcb7a35",
		Name:         "Ultimate Magic: Meteor",
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 7}},
		Completeness: CompletenessFull,
		SpecialActions: []game.SpecialAction{
			Foretell("{5}{R}"),
		},
		// The foretold choice waits for a prompt the damage asks (a CR 616 order;
		// CR 608.2c, #4155).
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error { return damageEachMatching(ctx, Creature(), 7) }),
			Instruction(func(ctx *Context) error {
				if ctx.Item.CastFromZone != game.ZoneExile {
					return nil
				}
				return ultimateMagicMeteorChoose(ctx, ctx.Item)
			}),
		),
	})
}

// ultimateMagicMeteorChoose walks the opponents, asking the caster for
// one artifact or land of each, then destroys the picks.
func ultimateMagicMeteorChoose(ctx *Context, item *game.StackItem) error {
	return ultimateMagicMeteorStep(ctx.Game, item, nil, ctx.Opponents())
}

// ultimateMagicMeteorStep asks about the first of the opponents still to
// walk who controls an artifact or a land, with the picks so far and the
// opponents after that one riding the prompt as its carry (the picks first,
// their count in CarryNums), and destroys the picks once none is left. The
// walk is data, not a closure (#4253).
func ultimateMagicMeteorStep(g *game.Game, item *game.StackItem, chosen, opps []uuid.UUID) error {
	for i := range opps {
		var options []uuid.UUID
		for _, c := range g.BattlefieldCardsForEffect() {
			if c.Controller == opps[i] && (c.IsArtifact() || c.IsLand()) {
				options = append(options, c.InstanceID)
			}
		}
		if len(options) == 0 {
			continue
		}
		carry := append(append([]uuid.UUID{}, chosen...), opps[i+1:]...)
		g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
			Chooser:   item.Controller,
			Source:    item.SourceCardID,
			Question:  "Ultimate Magic: Meteor — choose an artifact or land that player controls",
			Cards:     options,
			Min:       1,
			Max:       1,
			ThenKey:   ultimateMagicMeteorPicked,
			LiveZone:  game.ZoneBattlefield,
			Carry:     carry,
			CarryNums: []int{len(chosen)},
			Asked:     game.NewAskedItem(item),
		})
		return nil
	}
	dctx := NewContext(g, item).asGroupMember()
	for _, id := range chosen {
		if err := (DestroyTarget{Target: id}).Apply(dctx); err != nil {
			return err
		}
	}
	return nil
}

// ultimateMagicMeteorPicked is one opponent's pick, keyed (#4253): it joins
// the picks so far and walks on to the next opponent. Registered in init
// because the walk asks the next opponent, which names the key again.
var ultimateMagicMeteorPicked game.ChooseCardsThen

func init() {
	ultimateMagicMeteorPicked = askedPick("choose-cards/ultimate-magic-meteor-walk", func(ctx *Context, r game.CardsChosen) error {
		n := 0
		if len(r.Nums) > 0 {
			n = r.Nums[0]
		}
		if n < 0 || n > len(r.Carry) {
			return nil
		}
		chosen := append(append([]uuid.UUID{}, r.Carry[:n]...), r.Picked...)
		return ultimateMagicMeteorStep(ctx.Game, ctx.Item, chosen, r.Carry[n:])
	})
}
