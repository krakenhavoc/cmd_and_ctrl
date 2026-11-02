package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Wall Crawl — Enchantment {3}{G} (Marvel's Spider-Man, #4460):
//
//	"When this enchantment enters, create a 2/1 green Spider creature token
//	 with reach, then you gain 1 life for each Spider you control.
//	 Spiders you control get +1/+1 and can't be blocked by creatures with
//	 defender."
//
// The life is counted after the token exists (and after any prompt the
// creation raised), so the new Spider counts. Both static halves read the
// post-layer subtype, so a changeling counts as a Spider.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4a09536e-dfc1-4b15-adc1-07ef866ebda0",
		Name:         "Wall Crawl",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			TribalAnthem(TribeFilter{Tribes: []string{"Spider"}, YoursOnly: true}, 1, 1),
		},
		BlockRules: []game.BlockRule{
			CantBeBlockedBy(OnMatching(wallCrawlSpider),
				HasKeyword("defender"), "creatures with defender"),
		},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Wall Crawl — create a 2/1 Spider with reach, then gain 1 life for each Spider you control",
				Do(CreateToken{Template: TokenCard("2/1 green Spider with reach"), N: 1}, wallCrawlGainLife{})),
		},
	})
}

// wallCrawlGainLife is "then you gain 1 life for each Spider you control".
type wallCrawlGainLife struct{}

func (wallCrawlGainLife) Apply(ctx *Context) error {
	n := countControlled(ctx.Game, ctx.Controller(), func(c game.Card) bool {
		return c.IsCreature() && c.HasSubtypeWord("Spider")
	})
	return GainLife{Amount: n}.Apply(ctx)
}

// wallCrawlSpider is "Spiders you control" for the block rule, evaluated with
// the rule source's controller as the caster.
func wallCrawlSpider(_ *game.Game, caster uuid.UUID, c game.Card) bool {
	return c.Controller == caster && c.IsCreature() && c.HasSubtypeWord("Spider")
}
