package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Phoenix, Warden of Fire — the back face of Joshua, Phoenix's Dominant
// (Final Fantasy, #4462): Legendary Enchantment Creature — Saga Phoenix,
// 4/4.
//
//	"(As this Saga enters and after your draw step, add a lore counter.)
//	 I, II — Rising Flames — Phoenix deals 2 damage to each opponent.
//	 III — Flames of Rebirth — Return any number of target creature cards
//	 with total mana value 6 or less from your graveyard to the
//	 battlefield. Exile Phoenix, then return it to the battlefield (front
//	 face up).
//	 Flying, lifelink"
//
// A Saga creature. Chapters I and II are one damage event to every opponent.
// Chapter III targets any number of creature cards in the controller's
// graveyard with a total mana value of 6 or less (a card with {X} in its
// cost counts X as 0, CR 202.3e) and puts the ones still legal onto the
// battlefield together (The Rise of Sozin's reanimation); the Saga is then
// exiled and returned front face up as a new object (CR 712.14a). The flip
// does not depend on the reanimation: with no target the Saga is still
// exiled and returned.
//
// No simplification.
func init() {
	const name = "Phoenix, Warden of Fire"
	rising := Do(DamageEachOpponent{N: 2})
	Register(Spec{
		OracleID:        joshuaPhoenixsDominantOracleID + "#1",
		Name:            name,
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "lifelink"},
		Triggered: []game.TriggeredAbility{
			ChapterTrigger(1, SagaChapterLabel(name, 1, "Rising Flames — Phoenix deals 2 damage to each opponent"), rising),
			ChapterTrigger(2, SagaChapterLabel(name, 2, "Rising Flames — Phoenix deals 2 damage to each opponent"), rising),
			ChapterTriggerTargeting(3, SagaChapterLabel(name, 3, "Flames of Rebirth — return any number of creature cards with total mana value 6 or less from your graveyard to the battlefield; exile Phoenix, then return it front face up"),
				TargetCardInGraveyard("any number of target creature cards with total mana value 6 or less from your graveyard",
					Creature(), YouOwn()).WithCount(0, 0).WithTotal(TotalManaValueAtMost(6)),
				phoenixFlamesOfRebirth),
		},
	})
}
