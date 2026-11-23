package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Psemilla, Meletian Poet — Legendary Creature — Human Bard {2}{W}, 1/1
// (Foundations Jumpstart, #5929):
//
//	"Whenever you cast your first enchantment spell each turn, create a 2/2
//	 white Nymph enchantment creature token.
//	 At the beginning of each combat, if you control five or more
//	 enchantments, Psemilla gets +4/+4 and gains lifelink until end of
//	 turn."
//
// The first trigger reads the per-turn cast tally (CastTally.Enchantment,
// bumped before the cast event, so 1 means this spell is the first); an
// enchantment creature counts, and a copy or a spell cast by an opponent does
// not. The second is an intervening "if" (CR 603.4) on every player's
// beginning of combat, checked as the trigger fires and again as it
// resolves; the five enchantments are the ones you control then, Psemilla's
// own type not counting unless something made it an enchantment. The boost and
// lifelink are one effect of the same resolution (CR 611.2c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "12712419-a2a9-483d-a65a-2b727ae404ae",
		Name:         "Psemilla, Meletian Poet",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, AllOf(YouCast(Enchantment()), YouCastYourFirstEnchantmentSpellEachTurn),
				"Psemilla — create a 2/2 white Nymph enchantment creature token",
				Do(CreateToken{Template: TokenCard("2/2 white Nymph enchantment creature"), N: 1})),
			InterveningIf(psemillaFiveEnchantments,
				AtBeginningOfEachCombat("Psemilla — gets +4/+4 and gains lifelink until end of turn (you control five or more enchantments)",
					Do(Instruction(psemillaPump)))),
		},
	})
}

// youCastYourFirstEnchantmentSpellEachTurn is the "first ... each turn" half of
// the condition, as Arcane Bombardment's is for instants and sorceries.
func YouCastYourFirstEnchantmentSpellEachTurn(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return ev.Kind == game.EventCast && ev.Actor == source.Controller &&
		g.CastTallyFor(ev.Actor).Enchantment == 1
}

var psemillaFiveEnchantments = CondBoard{
	Name: "you control five or more enchantments",
	Test: func(ctx *Context) bool { return CountYouControl(ctx, Enchantment()) >= 5 },
}

func psemillaPump(ctx *Context) error {
	id := ctx.Item.SourceCardID
	if err := (BoostUntilEOT{Target: id, Power: 4, Toughness: 4, Label: "Psemilla — +4/+4"}).Apply(ctx); err != nil {
		return err
	}
	return GrantKeywordUntilEOT{Target: id, Keywords: []string{"lifelink"}, Label: "Psemilla — lifelink"}.Apply(ctx)
}
