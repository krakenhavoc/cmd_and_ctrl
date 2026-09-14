package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Carmen, Cruel Skymarcher — Legendary Creature — Vampire Soldier
// {3}{W}{B}, 2/2 (EDHREC rank 2806):
//
//	"Flying
//	 Whenever a player sacrifices a permanent, put a +1/+1 counter on
//	 Carmen and you gain 1 life.
//	 Whenever Carmen attacks, return up to one target permanent card
//	 with mana value less than or equal to Carmen's power from your
//	 graveyard to the battlefield."
//
// The aristocrats commander that grows on everyone's sacrifices and
// then reanimates on the attack. Two triggers:
//
//   - The sacrifice trigger is Mayhem Devil's: any player, any
//     permanent — a Treasure cracked for mana, a fetchland, an edict.
//     The counter lands only if Carmen is still on the battlefield;
//     the life is gained either way.
//   - The attack trigger targets a permanent card in the controller's
//     graveyard whose mana value is at most Carmen's CURRENT power,
//     read when the legal set is computed and again at resolution
//     (CR 608.2b) — so a counter placed in response widens the set,
//     and a shrink narrows it. "Up to one" is a Min-0 target; with
//     nothing chosen the ability does nothing. The card returns under
//     its owner's control, which is the controller's ("your
//     graveyard"), through the ordinary reanimation path.
//
// Sandbox simplification, declared: an Aura returned this way comes
// back unattached and stays that way — the reanimation path has no
// "choose what it enchants" prompt (the Brilliant Restoration
// posture). Weaker than printed, never stronger.
const b26CarmenOracle = "84c24e62-26bd-4ca4-b6b9-d31e2065f13b"

func init() {
	Register(Spec{
		OracleID:        b26CarmenOracle,
		Name:            "Carmen, Cruel Skymarcher",
		Completeness:    CompletenessCaveats,
		Caveats:         []string{"An Aura returned this way comes back unattached and stays that way — you don't get to choose what it enchants."},
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			{
				Watches: []game.EventKind{game.EventSacrifice},
				AppliesTo: func(ev game.Event, _ *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return ev.Kind == game.EventSacrifice && ev.CardID != uuid.Nil
				},
				Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
					return game.NewTriggeredItem(source, "Carmen, Cruel Skymarcher — +1/+1 counter and 1 life",
						b26PutCounterOnSelfAndGainLife)
				},
			},
			{
				Watches: []game.EventKind{game.EventAttack},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return attackDeclared(ev, source)
				},
				Targets: b26TargetPermanentCardInYourGraveyardWithManaValueAtMostPowerOf(b26CarmenOracle),
				Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
					return game.NewTriggeredItem(source, "Carmen, Cruel Skymarcher — return a permanent card to the battlefield",
						b26ReturnFirstLegalGraveyardTargetToBattlefield)
				},
			},
		},
	})
}
