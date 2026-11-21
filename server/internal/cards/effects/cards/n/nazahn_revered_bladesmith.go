package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nazahn, Revered Bladesmith — Legendary Creature — Cat Artificer
// {4}{G}{W}, 5/4 (Commander 2017, #5533):
//
//	"When Nazahn enters, search your library for an Equipment card and
//	 reveal it. If you reveal a card named Hammer of Nazahn this way, put
//	 it onto the battlefield. Otherwise, put that card into your hand. Then
//	 shuffle.
//	 Whenever an equipped creature you control attacks, you may tap target
//	 creature defending player controls."
//
// The search is Guidelight Pathmaker's: revealed, to the hand, and a found
// Hammer of Nazahn is then put onto the battlefield from it
// (c17ThenPutFoundHammerOntoBattlefield); not finding one is allowed
// (CR 701.23b), and the library is shuffled either way. The attack trigger
// is Swordsman, Sharp Scoundrel's (one per equipped attacker, "equipped"
// judged as the attack is declared) with Captain America's Shield's target:
// a creature the player that attacker is attacking controls. The tap is
// optional, asked as the trigger resolves.
//
// No simplification.
func init() {
	enters := effects.WhenThisEnters("Nazahn, Revered Bladesmith — search your library for an Equipment card; a Hammer of Nazahn onto the battlefield, otherwise into your hand",
		func(g *game.Game, item *game.StackItem) error {
			return effects.SearchLibrary{
				Player:    item.Controller,
				Predicate: effects.B09IsEquipmentCard,
				Dest:      game.ZoneHand,
				Limit:     1,
				Reveal:    true,
				Shuffle:   true,
				Reason:    "Nazahn, Revered Bladesmith — an Equipment card",
				ThenTail:  game.KeyedTail{Ref: effects.C17ThenPutFoundHammerOntoBattlefield, Player: item.Controller, Source: item.SourceCardID},
			}.Apply(effects.NewContext(g, item))
		})
	enters.Purpose = game.Purpose{Tutors: 1}
	attacks := effects.Optional(game.TriggeredAbility{
		Watches:     []game.EventKind{game.EventAttack},
		AppliesTo:   effects.EquippedCreatureYouControlAttacks,
		TargetsFrom: effects.MshB2TargetCreatureTheAttackerDefends,
		Key:         "Nazahn, Revered Bladesmith — you may tap target creature defending player controls",
		Effect:      effects.B36TapChosenCreature,
	}, "Nazahn, Revered Bladesmith — tap that creature?")
	effects.Register(effects.Spec{
		OracleID:     "a3a6c024-43de-4f5c-a1d6-c48a0c96a43b",
		Name:         "Nazahn, Revered Bladesmith",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{enters, attacks},
	})
}
