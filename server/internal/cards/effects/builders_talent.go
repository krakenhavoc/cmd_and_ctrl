package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Builder's Talent — Enchantment — Class {1}{W} (Bloomburrow, #4467):
//
//	"(Gain the next level as a sorcery to add its ability.)
//	 When this Class enters, create a 0/4 white Wall creature token with
//	 defender.
//	 {W}: Level 2
//	 Whenever one or more noncreature, nonland permanents you control
//	 enter, put a +1/+1 counter on target creature you control.
//	 {4}{W}: Level 3
//	 When this Class becomes level 3, return target noncreature, nonland
//	 permanent card from your graveyard to the battlefield."
//
// ADR 0071 Class machinery (Caretaker's Talent): LevelUp for each level,
// the level-2 trigger gated with AtLevel and the level-3 one a
// BecomesLevel with a target (CR 716.2a).
//
//   - Level 1 is an entry trigger making the Wall.
//   - Level 2 is Leader's Talent's targeted counter, on Caretaker's
//     Talent's "one or more … you control enter": once per batch, so a
//     board of artifacts entering at once is one counter. The entering
//     permanent's types are read as it entered.
//   - Level 3 returns the targeted card, if still legal (CR 608.2b), under
//     your control. Returning one is itself a noncreature, nonland
//     permanent entering, so level 2 sees it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0b19082c-8f80-448b-9e1b-0f324a159409",
		Name:         "Builder's Talent",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Builder's Talent — create a 0/4 white Wall creature token with defender",
				Do(CreateToken{Template: TokenCard("0/4 white Wall with defender"), N: 1})),
			AtLevel(2, OncePerBatch(Targeting(
				On(game.EventETB, buildersTalentNoncreatureNonlandEntered,
					"Builder's Talent — put a +1/+1 counter on target creature you control",
					Do(Instruction(counterOnFirstCreatureTarget))),
				TargetCreature("target creature you control", YouControl())))),
			Targeting(BecomesLevel(3, "Builder's Talent — return target noncreature, nonland permanent card from your graveyard to the battlefield",
				returnFirstLegalGraveyardTargetToBattlefield),
				TargetCardInGraveyard("target noncreature, nonland permanent card from your graveyard",
					YouOwn(), Permanent(), Noncreature(), Nonland())),
		},
		Activated: []ActivatedAbility{
			LevelUp(2, ManaCost("{W}")),
			LevelUp(3, ManaCost("{4}{W}")),
		},
	})
}

// buildersTalentNoncreatureNonlandEntered is "a noncreature, nonland
// permanent you control enters".
func buildersTalentNoncreatureNonlandEntered(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	c, ok := enteredUnderYourControl(ev, source, g, false)
	return ok && !c.IsCreature() && !c.IsLand()
}
