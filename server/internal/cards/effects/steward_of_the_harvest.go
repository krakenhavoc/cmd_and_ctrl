package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Steward of the Harvest — Creature — Human Druid {3}{G}, 3/3:
//
//	"When this creature enters, exile up to three target land cards
//	 from your graveyard.
//	 Creatures you control have all activated abilities of all land
//	 cards exiled with this creature."
//
// The enters exile is linked (ExileTarget.Linked, ADR 0155 §8), and the
// borrow lends to every creature you control, the Steward included
// (GrantActivatedAbilitiesOf over ExiledWithLenders). A land card in
// exile lends its printed abilities and, for a basic land type, its CR
// 305.6 mana ability, so an exiled Forest gives each of your creatures
// "{T}: Add {G}", usable only once the creature can tap (CR 302.6).
//
// No simplification.
func init() {
	const clause = "Creatures you control have all activated abilities of all land cards exiled with this creature."
	borrow := GrantActivatedAbilitiesOf(func(target *game.Card, _ *game.Game, source *game.Card) bool {
		return source != nil && target.IsCreature() && target.Controller == source.Controller
	}, game.AbilityBorrow{
		Label:   clause,
		Lenders: ExiledWithLenders(ByController(Land())),
	})
	borrow.DependsOnExile = true
	Register(Spec{
		OracleID:     "1bb7a942-c4b9-47a2-9782-3ec3efcdaff5",
		Name:         "Steward of the Harvest",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{borrow},
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventETB},
			AppliesTo: b06SelfETB,
			Targets:   TargetCardInGraveyard("up to three target land cards from your graveyard", Land(), YouOwn()).WithCount(0, 3),
			Key:       "Steward of the Harvest — exile up to three land cards from your graveyard",
			Effect:    exileLinkedTargetsEffect,
		}},
	})
}
