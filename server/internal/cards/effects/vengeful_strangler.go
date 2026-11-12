package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Vengeful Strangler // Strangling Grasp (Innistrad: Midnight Hunt,
// #5677).
//
// Front face, Creature — Human Rogue {1}{B}, 2/1:
//
//	"This creature can't block.
//	 When this creature dies, return it to the battlefield transformed
//	 under your control attached to target creature or planeswalker an
//	 opponent controls."
//
// Back face, Enchantment — Aura:
//
//	"Enchant creature or planeswalker an opponent controls
//	 At the beginning of your upkeep, enchanted permanent's controller
//	 sacrifices a nonland permanent of their choice, then that player
//	 loses 1 life."
//
// The dies trigger targets as it goes on the stack and returns the card
// from the graveyard back face up, already attached to the target
// (ReturnThisCardTransformedAttachedToTarget). The effect names the
// host, so nobody chooses one as it enters (CR 303.4f); with no legal
// target the trigger is removed and the card stays in the graveyard, as
// it does when the target has become illegal or the card has left the
// graveyard in the meantime. A token Strangler has no card to return.
//
// "An opponent controls" is part of the enchant ability, so the Aura is
// put into its owner's graveyard if its controller comes to control the
// enchanted permanent (CR 704.5m).
//
// The upkeep trigger reads the enchanted permanent's controller as it
// resolves. That player picks any nonland permanent of theirs, the
// enchanted one included, and loses 1 life once it has gone; a player
// with nothing to sacrifice still loses the life.
//
// No simplification.
func init() {
	const host = "creature or planeswalker an opponent controls"
	Register(Spec{
		OracleID:     vengefulStranglerOracleID,
		Name:         "Vengeful Strangler",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{RestrictSelf(game.CantBlock)},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisDies(
				"Vengeful Strangler — return it to the battlefield transformed attached to target "+host,
				ReturnThisCardTransformedAttachedToTarget),
				TargetPermanent("target "+host, Or(Creature(), Planeswalker()), OpponentControls())),
		},
	})
	Register(Spec{
		OracleID:     vengefulStranglerOracleID + "#1",
		Name:         "Strangling Grasp",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("enchant "+host, Or(Creature(), Planeswalker()), OpponentControls()),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Strangling Grasp — enchanted permanent's controller sacrifices a nonland permanent, then loses 1 life",
				stranglingGraspUpkeep),
		},
	})
}

const vengefulStranglerOracleID = "ffc1fe11-0ecb-4497-a266-0817c79477e9"

// stranglingGraspUpkeep is the upkeep trigger's resolution: the
// enchanted permanent's controller chooses and sacrifices a nonland
// permanent, and the life loss is the sacrifice's tail.
func stranglingGraspUpkeep(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	victim := stranglingGraspVictim(ctx)
	if p := g.PlayerByIDForEffect(victim); p == nil || p.Eliminated {
		return nil
	}
	return SacrificeChoice{
		Player:     victim,
		Candidates: NonlandPermanentsControlledBy(g, victim),
		Question:   "Strangling Grasp — sacrifice a nonland permanent",
		ThenTail:   game.KeyedTail{Ref: thenChangeLife, Player: victim, Source: item.SourceCardID, N: -1},
	}.Apply(ctx)
}

// stranglingGraspVictim is "enchanted permanent's controller" as the
// trigger resolves: the controller of the permanent the Aura is attached
// to, or was attached to as it last existed on the battlefield
// (CR 608.2h). uuid.Nil when it enchanted nothing.
func stranglingGraspVictim(ctx *Context) uuid.UUID {
	info, ok := ctx.SourcePermanent()
	if !ok || info.AttachedTo.Kind != game.TargetCard {
		return uuid.Nil
	}
	if z := ctx.Game.FindCardZoneForEffect(info.AttachedTo.ID); z != nil && z.Kind == game.ZoneBattlefield {
		if host, ok := ctx.Game.LookupCardForEffect(info.AttachedTo.ID); ok {
			return host.Controller
		}
	}
	if info.AttachedObject != nil {
		if host, ok := ctx.Game.PermanentForEffect(*info.AttachedObject); ok {
			return host.Controller
		}
	}
	return uuid.Nil
}
