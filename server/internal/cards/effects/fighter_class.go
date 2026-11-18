package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fighter Class — Enchantment — Class {R}{W} (Adventures in the Forgotten
// Realms):
//
//	"(Gain the next level as a sorcery to add its ability.)
//	 When this Class enters, search your library for an Equipment card,
//	 reveal it, put it into your hand, then shuffle.
//	 {1}{R}{W}: Level 2
//	 Equip abilities you activate cost {2} less to activate.
//	 {3}{R}{W}: Level 3
//	 Whenever a creature you control attacks, up to one target creature
//	 blocks it this combat if able."
//
// One line per level (ADR 0071):
//
//   - Level 1 is an entry trigger with Cloud, Midgar Mercenary's search for
//     an Equipment card (revealed, to hand, then shuffle), ungated (CR
//     716.3). The search is "a card": it may find nothing.
//   - Level 2 is Bureau Headmaster's activation discount on equip
//     abilities, for {2}, gated at level 2. The cost never goes below zero
//     (CR 601.2f).
//   - Level 3 is an attack trigger per creature you control that attacks,
//     gated at level 3, with "up to one" target creature. Its requirement
//     names the attacker the trigger came from as an object (instance and
//     epoch): a creature that left and came back is not it (CR 400.7).
//     It lasts until the end of the combat it was made in
//     (BlocksAttackerThisCombat), and a blocker that cannot legally block
//     that attacker is asked nothing (CR 509.1c).
//
// No simplification.
func init() {
	discount := EquipAbilitiesYouActivateCostLess(2, "Equip abilities you activate cost {2} less to activate.")
	discount.ActiveWhen = Level(2)
	Register(Spec{
		OracleID:      "8a280ba8-6b7b-46dd-9c2b-6d0027cfff30",
		Name:          "Fighter Class",
		Completeness:  CompletenessFull,
		Purpose:       game.Purpose{Tutors: 1},
		CostModifiers: []game.CostModifier{discount},
		Activated: []ActivatedAbility{
			LevelUp(2, ManaCost("{1}{R}{W}")),
			LevelUp(3, ManaCost("{3}{R}{W}")),
		},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Fighter Class — search your library for an Equipment card, reveal it, put it into your hand, then shuffle",
				func(g *game.Game, item *game.StackItem) error {
					return SearchLibrary{
						Player:    item.Controller,
						Predicate: b09IsEquipmentCard,
						Dest:      game.ZoneHand,
						Limit:     1,
						Reveal:    true,
						Shuffle:   true,
						Reason:    "Fighter Class — an Equipment card, revealed, to hand",
						Source:    item.SourceCardID,
					}.Apply(NewContext(g, item))
				}),
			AtLevel(3, Targeting(
				On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return attackDeclaredByYou(ev, source.Controller)
				}, "Fighter Class — up to one target creature blocks it this combat if able", fighterClassForceBlock),
				TargetCreature("up to one target creature").WithCount(0, 1))),
		},
	})
}

// fighterClassForceBlock makes the still-legal target block the creature that
// attacked, this combat, if able.
func fighterClassForceBlock(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	blocker, ok := b16FirstLegalTargetCard(ctx)
	if !ok || item.Trigger == nil {
		return nil
	}
	attacker := item.Trigger.Event.CardID
	if !onBattlefield(g, attacker) {
		return nil
	}
	ref, ok := g.PermanentRefForEffect(attacker)
	if !ok {
		return nil
	}
	return BlocksAttackerThisCombat{
		Blocker:  blocker,
		Attacker: ref,
		Label:    "Fighter Class — blocks the attacking creature this combat if able",
	}.Apply(ctx)
}
