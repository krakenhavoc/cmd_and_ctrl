package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Esper Terra — the back face of Terra, Magical Adept (Final Fantasy,
// #4462): Legendary Enchantment Creature — Saga Wizard, 6/6.
//
//	"(As this Saga enters and after your draw step, add a lore counter.)
//	 I, II, III — Create a token that's a copy of target nonlegendary
//	 enchantment you control. It gains haste. If it's a Saga, put up to
//	 three lore counters on it. Sacrifice it at the beginning of your next
//	 end step.
//	 IV — Add {W}{W}, {U}{U}, {B}{B}, {R}{R}, and {G}{G}. Exile Esper
//	 Terra, then return it to the battlefield (front face up).
//	 Flying"
//
// A Saga creature (ChapterTriggerTargeting). Chapters I–III are the token
// copy of the targeted enchantment, made through CreateTokenCopiesTogether
// whose continuation, told the token, gives it haste (a continuous effect
// with no duration), offers the Saga's lore counters through a chosen
// number from 0 to 3 (each counter triggers its chapter as it is put on,
// CR 714.2c), and schedules the sacrifice at its creator's next end step
// (CR 603.7). Chapter IV adds the ten mana in one go and flickers the Saga,
// which returns front face up as a new object (CR 712.14a).
//
// No simplification.
func init() {
	const name = "Esper Terra"
	target := func() *game.TargetSpec {
		return TargetPermanent("target nonlegendary enchantment you control", Enchantment(), Not(Legendary()), YouControl())
	}
	label := func(n int) string {
		return SagaChapterLabel(name, n, "create a token copy of target nonlegendary enchantment you control; it gains haste; sacrifice it at your next end step")
	}
	Register(Spec{
		OracleID:        terraMagicalAdeptOracleID + "#1",
		Name:            name,
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			ChapterTriggerTargeting(1, label(1), target(), esperTerraCopy),
			ChapterTriggerTargeting(2, label(2), target(), esperTerraCopy),
			ChapterTriggerTargeting(3, label(3), target(), esperTerraCopy),
			ChapterTrigger(4, SagaChapterLabel(name, 4, "add {W}{W}, {U}{U}, {B}{B}, {R}{R}, and {G}{G}; exile Esper Terra, then return it front face up"),
				esperTerraManaAndReturn),
		},
	})
}
