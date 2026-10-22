package effects

import (
	"slices"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Command the Stage — Sorcery {2}{R}:
//
//	"Create a 2/2 colorless Wizard Soldier creature token named Cadet,
//	 then put a +1/+1 counter on each other Wizard token you control.
//	 At the beginning of each upkeep, if an opponent was dealt
//	 noncombat damage last turn, return this card from your graveyard
//	 to your hand."
//
// The spell is whole: the Cadet is created first, then every OTHER
// Wizard token its controller has (the new Cadet is excluded by
// instance ID) gets a counter, through AddCounter so a counter
// replacement applies.
//
// The graveyard-recursion trigger (#3172) is a trigger that works from the
// graveyard (TriggeredAbility.Zones) on each upkeep. Its intervening if
// reads the previous turn's record of who was dealt noncombat damage
// (Game.WasDealtNoncombatDamageLastTurn, stamped as the turn rolled over),
// for any opponent of the card's owner, and the return re-checks that the
// card is still in the graveyard.
func init() {
	Register(Spec{
		OracleID:     "13bfc51f-d079-40ef-a4ba-47d06d3150a3",
		Name:         "Command the Stage",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{commandTheStageReturn()},
		// The counters wait for a prompt the Cadet's creation asks (a CR 616 order; CR 608.2c,
		// #4155) and skip the token it made, read from the event log after the cursor the
		// first step filed.
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				ctx.RecordStep(commandTheStageRecord, nil, int(b25LastEventSeq(ctx.Game)))
				_, err := ctx.Game.CreateTokensForEffect(ctx.Controller(), ctx.TextToken(TokenCard("2/2 colorless Wizard Soldier named Cadet")), 1, game.TokenEntryOptions{})
				return err
			}),
			Instruction(func(ctx *Context) error {
				me := ctx.Controller()
				made := b27TokensCreatedByAfter(ctx.Game, me, uint64(ctx.StepRecord(commandTheStageRecord).N()))
				// "Wizard" is a word of the spell's text (CR 612.2, #3984).
				wizard := ctx.TextSubtype("Wizard")
				for _, c := range ctx.Game.BattlefieldCardsForEffect() {
					if c.Controller != me || !IsToken(c) || !c.HasSubtype(wizard) || slices.Contains(made, c.InstanceID) {
						continue
					}
					if err := (AddCounter{Target: c.InstanceID, Kind: game.CounterPlusOne, N: 1}).Apply(ctx.asGroupMember()); err != nil {
						return err
					}
				}
				return nil
			}),
		),
	})
}

// commandTheStageReturn is "At the beginning of each upkeep, if an
// opponent was dealt noncombat damage last turn, return this card from
// your graveyard to your hand."
func commandTheStageReturn() game.TriggeredAbility {
	t := On(game.EventBeginUpkeep, func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		for _, p := range g.Seats {
			if p == nil || p.Eliminated || p.ID == source.Owner {
				continue
			}
			if g.WasDealtNoncombatDamageLastTurn(p.ID) {
				return true
			}
		}
		return false
	}, "Command the Stage — return this card from your graveyard to your hand", returnThisCardFromYourGraveyardToYourHand)
	t.Zones = []game.ZoneKind{game.ZoneGraveyard}
	return t
}

// commandTheStageRecord is the last event logged before the Cadet was made.
const commandTheStageRecord = "command-the-stage/before-cadet"
