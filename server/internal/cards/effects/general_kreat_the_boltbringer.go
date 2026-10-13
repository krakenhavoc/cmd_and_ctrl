package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// General Kreat, the Boltbringer — Legendary Creature — Goblin
// Soldier {2}{R}, 2/2 (EDHREC rank 2289):
//
//	"Whenever one or more Goblins you control attack, create a 1/1 red
//	 Goblin creature token that's tapped and attacking.
//	 Whenever another creature you control enters, General Kreat deals
//	 1 damage to each opponent."
//
// The Goblin general that makes a Goblin and pings for it. Two
// abilities, and they chain: the token the first makes is another
// creature entering, so the second fires and each opponent takes
// one — as printed.
//
//   - "One or more Goblins you control attack" is ONE trigger per
//     attack declaration. The engine emits EventAttack per creature,
//     so the AppliesTo declines any further Goblin's event from the
//     same batch — one declaration is one batch (OncePerBatch, see
//     docs/adding-cards.md) — keyed by label, because the ping trigger is a
//     different ability and must not swallow it. Kreat is a Goblin
//     and counts for his own trigger; effective subtypes, so a
//     changeling counts.
//   - The token enters TAPPED AND ATTACKING through
//     CreateTokensAttackingForEffect (Parhelion II's path): put onto
//     the battlefield attacking, never declared, so it fires no
//     "whenever a creature attacks" trigger of its own — CR 508.4 —
//     and it can be blocked and deals its 1 this combat.
//   - The ping is b13AnotherCreatureYouControlEntered: any other
//     creature, token or not, dealt by Kreat so it is red noncombat
//     damage.
//
// The controller chooses what the token attacks (CR 508.4) through
// CreateTokensAttackingYourChoice, asking about what the first Goblin
// attacked first. No simplification.
func init() {
	Register(Spec{
		OracleID:     "da1c1c70-c2fa-4ba7-89fe-d9af3fb353b9",
		Name:         "General Kreat, the Boltbringer",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			{
				OncePerBatch: true,
				Watches:      []game.EventKind{game.EventAttack},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return b21CreatureOfSubtypeYouControlAttacked(ev, source, g, "Goblin")
				},
				Key: b21KreatAttackLabel,
				Effect: func(g *game.Game, item *game.StackItem) error {
					return CreateTokensAttackingYourChoice{
						Template: b21TappedAttackingGoblin(),
						N:        1,
						Tapped:   true,
						Prefer:   triggerAttackTarget(item),
						Label:    "General Kreat",
					}.Apply(NewContext(g, item))
				},
			},
			WheneverAnotherCreatureEntersUnderYourControl("General Kreat — 1 damage to each opponent", Do(DamageEachOpponent{N: 1})),
		},
	})
}

// b21KreatAttackLabel is the attack trigger's stack label — named
// because the "one or more" dedup matches on it.
const b21KreatAttackLabel = "General Kreat — a tapped and attacking Goblin"
