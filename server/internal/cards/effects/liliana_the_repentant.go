package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Liliana the Repentant — Legendary Creature — Human Warlock {1}{B},
// 2/2 (Reality Fracture):
//
//	"Whenever another creature or planeswalker you control enters, mill
//	 two cards.
//	 Exhaust — {5}{B}: Return target creature or planeswalker card from
//	 your graveyard to the battlefield. Put a +1/+1 counter on Liliana.
//	 Activate only as a sorcery. (Activate each exhaust ability only
//	 once.)"
//
// The counter is not conditional on the return ("Put a +1/+1 counter on
// Liliana" is its own sentence), so it lands even if the target card
// left the graveyard in response — but then the whole ability has lost
// its only target and does nothing (CR 608.2b), exactly as printed.
func init() {
	Register(Spec{
		OracleID:     "5eb4403f-f199-4f75-a7c6-e76783f9b07d",
		Name:         "Liliana the Repentant",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, anotherCreatureOrPlaneswalkerEnteredUnderYourControl,
				"Liliana the Repentant — mill two cards",
				Do(MillCards{N: 2})),
		},
		Activated: []ActivatedAbility{{
			Label:        "Exhaust — {5}{B}: Return target creature or planeswalker card from your graveyard to the battlefield. Put a +1/+1 counter on Liliana. Activate only as a sorcery.",
			Exhaust:      true,
			Cost:         ManaCost("{5}{B}"),
			SorcerySpeed: true,
			Targets:      TargetCardInGraveyard("target creature or planeswalker card from your graveyard", YouOwn(), Or(Creature(), Planeswalker())),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				legal := ctx.LegalTargets()
				if len(legal) == 0 {
					return nil
				}
				for _, t := range legal {
					if t.Kind != game.TargetCard {
						continue
					}
					if err := (ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneBattlefield, Controller: item.Controller}).Apply(ctx); err != nil {
						return err
					}
				}
				return plusOneCountersOnThis(1)(g, item)
			},
		}},
	})
}

// anotherCreatureOrPlaneswalkerEnteredUnderYourControl is "whenever
// another creature or planeswalker you control enters".
func anotherCreatureOrPlaneswalkerEnteredUnderYourControl(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	c, ok := enteredUnderYourControl(ev, source, g, true)
	return ok && (c.IsCreature() || c.IsPlaneswalker())
}
