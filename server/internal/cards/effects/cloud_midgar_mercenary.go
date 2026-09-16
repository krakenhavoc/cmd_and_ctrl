package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cloud, Midgar Mercenary — Legendary Creature — Human Soldier
// Mercenary {W}{W}, 2/1 (EDHREC rank 1607):
//
//	"When Cloud enters, search your library for an Equipment card,
//	 reveal it, put it into your hand, then shuffle.
//	 As long as Cloud is equipped, if a triggered ability of Cloud or
//	 an Equipment attached to it triggers, that ability triggers an
//	 additional time."
//
// Stoneforge Mystic as a commander. The ETB is the Equipment tutor
// — mandatory here, unlike the Mystic's "you may", so the prompt
// opens every time and "fail to find" is still a legal answer (CR
// 701.23b).
//
// Sandbox simplification, declared — one whole ability omitted, the
// Stoneforge Mystic posture: the trigger-doubling static is NOT
// implemented. "That ability triggers an additional time" is a rule
// about how the trigger harvester queues abilities (Panharmonicon's
// family), and nothing in the engine lets a permanent multiply
// another permanent's triggers. Cloud's own only trigger is the ETB,
// which can never fire while he is equipped (he has just entered), so
// the whole of the gap is the attached Equipment's triggers — a Sword
// hitting once instead of twice. Weaker than printed, never stronger,
// and the caveat says so.
func init() {
	Register(Spec{
		OracleID:     "33d2584b-bf29-4c22-bd45-14ba2fb98c0e",
		Name:         "Cloud, Midgar Mercenary",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"The second ability isn't implemented — triggered abilities of Equipment attached to Cloud don't trigger an additional time."},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Cloud, Midgar Mercenary — search for an Equipment card", func(g *game.Game, item *game.StackItem) error {
				return SearchLibrary{
					Player:    item.Controller,
					Predicate: b09IsEquipmentCard,
					Dest:      game.ZoneHand,
					Limit:     1,
					Reveal:    true,
					Shuffle:   true,
					Reason:    "Cloud, Midgar Mercenary — an Equipment card, revealed, to hand",
				}.Apply(NewContext(g, item))
			}),
		},
	})
}
