package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ixhel, Scion of Atraxa — Legendary Creature — Phyrexian Angel
// {1}{W}{B}{G}, 2/5 (Phyrexia: All Will Be One Commander):
//
//	"Flying, vigilance, toxic 2
//	 Corrupted — At the beginning of your end step, each opponent who has
//	 three or more poison counters exiles the top card of their library face
//	 down. You may look at and play those cards for as long as they remain
//	 exiled, and you may spend mana as though it were mana of any color to
//	 cast those spells."
//
// Not an intervening "if": the trigger fires every end step, and each
// opponent is judged as it resolves (CR 608.2), so one who has fewer than
// three poison counters then exiles nothing. The cards are exiled face down
// (game.FaceDownPermitted): only their holder, Ixhel's controller, may look
// at them, and each is granted a play permission (lands included, "play") for
// as long as it stays in exile, with any-color mana for casting it
// (CastPermission.AnyColor; mana of any TYPE is not printed, so {C} costs
// stay as they are). An opponent with an empty library exiles nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "9fb7982f-fe35-4dea-a515-a0039ddbbaec",
		Name:            "Ixhel, Scion of Atraxa",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance", "toxic 2"},
		Triggered: []game.TriggeredAbility{
			effects.AtYourEndStep("Ixhel, Scion of Atraxa — each opponent with three or more poison counters exiles the top card of their library face down; you may play them",
				ixhelCorruptedExile),
		},
	})
}

// ixhelCorruptedExile is the end-step trigger: each opponent who has three
// or more poison counters, now, exiles their top card face down for
// Ixhel's controller to play.
func ixhelCorruptedExile(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	for _, opp := range ctx.Opponents() {
		p := g.PlayerByIDForEffect(opp)
		if p == nil || p.Eliminated || p.Poison < 3 {
			continue
		}
		if err := (effects.ExileTopWithPermission{
			From:        opp,
			GrantTo:     item.Controller,
			N:           1,
			AnyColor:    true,
			FaceDown:    true,
			WhileExiled: true,
		}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
