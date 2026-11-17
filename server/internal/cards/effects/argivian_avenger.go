package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Argivian Avenger — Artifact Creature — Shapeshifter {6}, 5/5 (The
// Brothers' War, #5482):
//
//	"{1}: Until end of turn, this creature gets -1/-1 and gains your
//	 choice of flying, vigilance, deathtouch, or haste."
//
// The shrink applies as the ability resolves, then the keyword is chosen
// (Lunar Avenger's pick). An Avenger that left in response gets nothing
// (CR 400.7); one that died to the -1/-1 gains nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d215cd4a-412c-4632-ab2d-de51103e6a7d",
		Name:         "Argivian Avenger",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{1}: Until end of turn, this creature gets -1/-1 and gains your choice of flying, vigilance, deathtouch, or haste.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:    ManaCost("{1}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !sourceIsStillThisPermanent(g, item) {
					return nil
				}
				if err := thisGetsUntilEndOfTurn(-1, -1, "Argivian Avenger — -1/-1 until end of turn")(g, item); err != nil {
					return err
				}
				return PickOption{
					Question: "Argivian Avenger gains which until end of turn?",
					Options: []game.ChoiceOption{
						{Label: "Flying"}, {Label: "Vigilance"}, {Label: "Deathtouch"}, {Label: "Haste"},
					},
					ThenKey: argivianAvengerAnswer,
				}.Apply(NewContext(g, item))
			},
		}},
	})
}

// argivianAvengerKeywords are the four choices, in the order the question lists them.
var argivianAvengerKeywords = []string{"flying", "vigilance", "deathtouch", "haste"}

// argivianAvengerAnswer grants the keyword chosen to the Avenger (the
// prompt's source) until end of turn. An on-disk key: never renamed.
var argivianAvengerAnswer = OptionPickThen("option-pick/s82-bro-b11-argivian-avenger-keyword", func(ctx *Context, r game.OptionPicked) error {
	if r.Index < 0 || r.Index >= len(argivianAvengerKeywords) {
		return nil
	}
	return GrantKeywordUntilEOT{Target: r.Source, Keywords: []string{argivianAvengerKeywords[r.Index]}}.Apply(ctx)
})
