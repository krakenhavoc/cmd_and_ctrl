package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sidequest: Play Blitzball // World Champion, Celestial Weapon —
// Enchantment // Legendary Artifact — Equipment {2}{R} (Final Fantasy,
// #4462):
//
//	Front: "At the beginning of combat on your turn, target creature you
//	 control gets +2/+0 until end of turn.
//	 At the end of combat on your turn, if a player was dealt 6 or more
//	 combat damage this turn, transform this enchantment, then attach it
//	 to a creature you control."
//	Back: "Double Overdrive — Equipped creature gets +2/+0 and has double
//	 strike.
//	 Equip {3}"
//
// The front is Storm the Vault's transforming enchantment. The combat
// trigger targets as it goes on the stack and gives the still-legal creature
// +2/+0. The end-of-combat ability has an intervening "if" (CR 603.4): it
// triggers only if some player has been dealt 6 or more combat damage this
// turn in total, from any sources (the recipient's per-turn tally, damage
// actually dealt), and that is checked again as it resolves. It transforms
// in place (TransformThis, CR 701.27), and then the Equipment, now an
// Equipment, is attached to a creature you control chosen as the ability
// resolves (not a target); with no creature it stays unattached. The back
// face is registered under "<oracle_id>#1" and is Dragoon's Lance's
// pump-and-keyword Equipment. Double Overdrive is an ability word.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     sidequestPlayBlitzballOracleID,
		Name:         "Sidequest: Play Blitzball",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				AtBeginningOfYourCombat("Sidequest: Play Blitzball — target creature you control gets +2/+0 until end of turn",
					func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						id, ok := b16FirstLegalTargetCard(ctx)
						if !ok {
							return nil
						}
						return BoostUntilEOT{Target: id, Power: 2, Label: "Sidequest: Play Blitzball — +2/+0"}.Apply(ctx)
					}),
				TargetCreature("target creature you control", YouControl())),
			On(game.EventStepBegan, AllOf(StepBegan(game.StepEndCombat, true),
				func(_ game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
					return aPlayerWasDealtSixCombatDamage(g)
				}),
				"Sidequest: Play Blitzball — transform it, then attach it to a creature you control",
				func(g *game.Game, item *game.StackItem) error {
					// CR 603.4's second check, as the ability resolves.
					if !aPlayerWasDealtSixCombatDamage(g) {
						return nil
					}
					ctx := NewContext(g, item)
					if err := (TransformThis{}).Apply(ctx); err != nil {
						return err
					}
					return sidequestBlitzballAttach(ctx)
				}),
		},
	})
}

// sidequestPlayBlitzballOracleID is shared with the back face, which
// registers under it plus "#1" (game.CatalogKey).
const sidequestPlayBlitzballOracleID = "9c0255ff-6579-4ce4-810d-01df9922ab33"

// aPlayerWasDealtSixCombatDamage is "a player was dealt 6 or more combat
// damage this turn".
func aPlayerWasDealtSixCombatDamage(g *game.Game) bool {
	for _, p := range g.Seats {
		if p != nil && g.CombatDamageTakenThisTurn(p.ID) >= 6 {
			return true
		}
	}
	return false
}

// sidequestBlitzballAttach is "then attach it to a creature you control":
// the pick of one creature, when the source really turned over and is the
// Equipment now.
func sidequestBlitzballAttach(ctx *Context) error {
	info, ok := ctx.SourcePermanent()
	if !ok || info.Left {
		return nil
	}
	src, ok := ctx.Game.LookupCardForEffect(ctx.Source())
	if !ok || !src.HasSubtype("Equipment") {
		return nil
	}
	return ChoosePermanents{
		Question:   "Sidequest: Play Blitzball — attach the Equipment to a creature you control",
		Keyed:      sidequestBlitzballAttached,
		Carry:      []uuid.UUID{ctx.Source()},
		Candidates: finB3CreaturesYouControl(1, 1),
	}.Apply(ctx)
}

// sidequestBlitzballAttached is the chosen creature wearing the Equipment,
// keyed (#4290).
var sidequestBlitzballAttached = RegisterPermanentsPick("prompt-run/sidequest-play-blitzball-attach", PermanentsDef{
	ThenRun: func(ctx *Context, picked game.PromptedPicks, r game.RunDone) error {
		ids := picked.Cards()
		if len(ids) == 0 || len(r.Carry) == 0 {
			return nil
		}
		return ctx.Game.AttachForEffect(r.Carry[0], game.TargetRef{Kind: game.TargetCard, ID: ids[0]})
	},
})
