package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Captain Jonathan Archer — Legendary Creature — Human Officer {2}{G}, 3/3
// (Star Trek, #4973):
//
//	"When Captain Archer enters, create Porthos, a legendary 1/1 green Dog
//	 creature token.
//	 Federation — {G}, {T}: Target creature you control gets +X/+X until
//	 end of turn, where X is the number of creature types among non-Borg
//	 creatures you control. Activate only as a sorcery."
//
// Porthos is the printed ttrk token, legendary, so a second one is subject
// to the legend rule. "Federation" is an ability word (CR 207.2c). X is
// Hoshi Sato's count (trekB7CreatureTypesAmongNonBorgYouControl), taken
// once as the ability resolves (CR 608.2h): each creature type among the
// non-Borg creatures the controller controls counts once, a changeling
// is a Borg and counts for nothing. A target gone by then gets nothing
// (CR 608.2b). The {T} keeps summoning sickness in force.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "85df3564-5134-4dc2-9021-40ec84ee6d08",
		Name:         "Captain Jonathan Archer",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Captain Jonathan Archer — create Porthos, a legendary 1/1 green Dog",
				Do(CreateToken{Template: TokenCard("1/1 green Legendary Dog named Porthos"), N: 1})),
		},
		Activated: []ActivatedAbility{{
			Label:        "Federation — {G}, {T}: Target creature you control gets +X/+X until end of turn, where X is the number of creature types among non-Borg creatures you control. Activate only as a sorcery.",
			Purpose:      ForTargets(BenefitUntilEOT(0, 0, 0)),
			Cost:         Plus(ManaCost("{G}"), TapCost()),
			Targets:      TargetCreature("target creature you control", YouControl()),
			SorcerySpeed: true,
			Effect: func(g *game.Game, item *game.StackItem) error {
				x := trekB7CreatureTypesAmongNonBorgYouControl(g, item)
				ctx := NewContext(g, item)
				for _, id := range legalTargetCards(item, g) {
					if err := (BoostUntilEOT{
						Target: id, Power: x, Toughness: x,
						Label: "Captain Jonathan Archer — +X/+X until end of turn",
					}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			},
		}},
	})
}
