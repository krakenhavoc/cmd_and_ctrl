package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Shigeki, Jukai Visionary — {1}{G} Legendary Enchantment Creature —
// Snake Druid:
//
//	"{1}{G}, {T}, Return Shigeki to its owner's hand: Reveal the top
//	 four cards of your library. You may put a land card from among
//	 them onto the battlefield tapped. Put the rest into your graveyard.
//	 Channel — {X}{X}{G}{G}, Discard this card: Return X target
//	 nonlegendary cards from your graveyard to your hand."
//
// #2028: the return is the cost (ReturnThis), paid with the mana and
// the tap at announce (CR 602.2b), so Shigeki is back in its owner's
// hand before the dig resolves and can be cast again. As a commander,
// its owner is asked whether it goes to the command zone instead
// (CR 903.9b), before anything is paid, as every commander returned as
// a cost is (ADR 0115).
//
// Channel is Otawara's shape: an ability that functions from the hand
// (Zones) with the discard-this cost. Its X targets are counted by the
// announced X (targetsCountedByX), which the {X}{X} in the cost makes
// cost two each.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "cbad3570-5417-42e6-b3ef-f42194098314",
		Name:         "Shigeki, Jukai Visionary",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:  "{1}{G}, {T}, Return Shigeki to its owner's hand: Reveal the top four cards of your library. You may put a land card from among them onto the battlefield tapped. Put the rest into your graveyard.",
				Cost:   Plus(ManaCost("{1}{G}"), TapCost(), ReturnThis()),
				Effect: shigekiDig,
			},
			{
				Label: "Channel — {X}{X}{G}{G}, Discard this card: Return X target nonlegendary cards from your graveyard to your hand.",
				Cost:  game.AbilityCost{Mana: "{X}{X}{G}{G}", DiscardSelf: true},
				Zones: []game.ZoneKind{game.ZoneHand},
				Targets: targetsCountedByX(TargetCardInGraveyard("X target nonlegendary cards from your graveyard",
					YouOwn(), Not(Legendary()))),
				Effect: returnEachTargetCardToHand,
			},
		},
	})
}
