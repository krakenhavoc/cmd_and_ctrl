package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wingnut, Bat on the Belfry — {1}{R} Legendary Creature — Bat Mutant 1/2
// (Teenage Mutant Ninja Turtles, #4457):
//
//	"Alliance — Whenever another creature you control enters, Wingnut gains
//	 your choice of flying, menace, or haste until end of turn.
//	 Whenever Wingnut attacks, each other attacking creature gets +1/+0
//	 until end of turn."
//
// Alliance is an ability word (CR 207.2c) on the "another creature you
// control enters" trigger. The keyword is chosen as the trigger resolves,
// Lunar Avenger's prompt, and is granted only while Wingnut is still the
// same permanent. The attack trigger is Hero of Bladehold's battle cry
// spelled out: the set of other attacking creatures is fixed as it
// resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "392a71c7-8f6d-46b8-bcc3-613e086d86ed",
		Name:         "Wingnut, Bat on the Belfry",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverAnotherCreatureEntersUnderYourControl("Wingnut — alliance: gains your choice of flying, menace, or haste until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					if !sourceIsStillThisPermanent(g, item) {
						return nil
					}
					return PickOption{
						Question: "Wingnut gains which until end of turn?",
						Options:  []game.ChoiceOption{{Label: "Flying"}, {Label: "Menace"}, {Label: "Haste"}},
						ThenKey:  wingnutKeywordChosen,
					}.Apply(NewContext(g, item))
				}),
			WheneverThisAttacks("Wingnut — each other attacking creature gets +1/+0 until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					return BoostUntilEOT{
						Match: And(AttackingCreature(), OtherThan(item.SourceCardID)),
						Power: 1,
						Label: "Wingnut — +1/+0 to each other attacking creature",
					}.Apply(NewContext(g, item))
				}),
		},
	})
}

// wingnutKeywords are the three choices, in the order the question lists them.
var wingnutKeywords = []string{"flying", "menace", "haste"}

// wingnutKeywordChosen grants the chosen keyword to Wingnut (the prompt's
// source) until end of turn, keyed (#4253).
var wingnutKeywordChosen = OptionPickThen("option-pick/wingnut-bat-on-the-belfry-keyword", func(ctx *Context, r game.OptionPicked) error {
	if r.Index < 0 || r.Index >= len(wingnutKeywords) || !onBattlefield(ctx.Game, r.Source) {
		return nil
	}
	return GrantKeywordUntilEOT{
		Target:   r.Source,
		Keywords: []string{wingnutKeywords[r.Index]},
		Label:    "Wingnut — " + wingnutKeywords[r.Index] + " until end of turn",
	}.Apply(ctx)
})
