package effects

import (
	"strconv"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Armor Wars — Enchantment — Saga {2}{U}{R} (Marvel Super Heroes, #4455):
//
//	"(As this Saga enters and after your draw step, add a lore counter.
//	 Sacrifice after III.)
//	 I — You may draw a card for each artifact you control. If you do,
//	 each opponent draws a card.
//	 II — Artifact spells you cast this turn cost {1} less to cast.
//	 III — This Saga deals X damage to target opponent, where X is the
//	 greatest mana value among artifacts you control."
//
// I counts the artifacts as it resolves and asks once: yes draws that many
// cards and then each opponent draws one. With no artifact there is
// nothing to draw, so "if you do" is never met and no one draws; a player
// a draw limit stops from drawing that many cannot choose to (CR 121.2b,
// MayChoice.Draws). II is Urza's cost promise (GrantNextSpellPromise with
// EveryMatchingSpell) over artifact spells for the turn. III reads the
// greatest mana value at resolution, so an artifact destroyed in response
// to the chapter does not count.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1ad83cac-2c85-477b-ab00-1369f9683db5",
		Name:         "Armor Wars",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			ChapterTrigger(1, SagaChapterLabel("Armor Wars", 1, "you may draw a card for each artifact you control; if you do, each opponent draws a card"),
				armorWarsDraw),
			ChapterTrigger(2, SagaChapterLabel("Armor Wars", 2, "artifact spells you cast this turn cost {1} less"),
				Do(GrantNextSpellPromise{From: "Armor Wars", Promise: game.NextSpellPromise{
					Filter:             game.PermissionFilter{CardType: "artifact"},
					Reduce:             1,
					EveryMatchingSpell: true,
					Text:               "Artifact spells you cast this turn cost {1} less to cast.",
				}})),
			ChapterTriggerTargeting(3, SagaChapterLabel("Armor Wars", 3, "damage to target opponent equal to the greatest mana value among your artifacts"),
				TargetPlayer("target opponent", Opponent()), armorWarsDamage),
		},
	})
}

// armorWarsDraw is chapter I: the question, with the artifact count it
// would draw.
func armorWarsDraw(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	n := b03ArtifactsControlled(g, item.Controller)
	if n <= 0 {
		return nil
	}
	return MayChoice{
		Question: "Armor Wars — draw " + strconv.Itoa(n) + " cards (one for each artifact you control)? If you do, each opponent draws a card.",
		YesLabel: "Draw, and each opponent draws a card",
		NoLabel:  "Don't draw",
		Draws:    n,
		Word:     strconv.Itoa(n),
		Then:     armorWarsDrawAnswer,
	}.Apply(ctx)
}

// armorWarsDrawAnswer is the yes: the cards for the artifacts, then one
// for each opponent.
var armorWarsDrawAnswer = MayYes("confirm/armor-wars-draw", func(ctx *Context, r game.ConfirmAnswer) error {
	if err := (DrawCards{Player: r.Chooser, N: wordInt(r)}).Apply(ctx); err != nil {
		return err
	}
	for _, opp := range ctx.Opponents() {
		if err := (DrawCards{Player: opp, N: 1}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
})

// armorWarsDamage is chapter III: the Saga deals the damage.
func armorWarsDamage(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		return DealDamage{
			Source: item.SourceCardID,
			Target: t.ID,
			Amount: b27GreatestManaValueAmongArtifactsControlled(g, item.Controller),
		}.Apply(ctx)
	}
	return nil
}
