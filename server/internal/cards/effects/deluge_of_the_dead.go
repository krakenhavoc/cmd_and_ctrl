package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Deluge of the Dead — Enchantment, the BACK FACE of Invasion of
// Innistrad (oracle a3c1af66…, face 1):
//
//	"When this enchantment enters, create two 2/2 black Zombie
//	 creature tokens.
//	 {2}{B}: Exile target card from a graveyard. If it was a creature
//	 card, create a 2/2 black Zombie creature token."
//
// # Why this file exists now and not in S27
//
// The front face shipped in #415 with the largest declared
// simplification in that sprint: the defeated Siege was exiled and
// its back face was never cast, because nothing in the multi-face
// model could say "cast THAT FACE of this exiled card". S32 closes
// that — ExilePlayPermission.Face, faceForCastLocked and
// faceOnResolve — so the back face finally becomes reachable and
// finally needs rules.
//
// # Why it registers under "<oracle_id>#1"
//
// Scryfall issues one oracle_id per CARD, so the battle and the
// enchantment share one, and effects.Register panics on a duplicate.
// game.CatalogKey makes the key composite: face 0 keeps the bare
// oracle ID and face N takes "<oracle_id>#N". Same keyspace the
// sixty MDFC land backs use (mdfc_lands.go); this is the first
// SPELL back face in it, and the first one reached by casting rather
// than by playing a land.
//
// The ETB is a real triggered ability rather than an OnETB hook,
// matching the front face: it uses the stack and can be responded to,
// which is what the card prints. The activated ability is Scavenging
// Ooze's shape exactly — read the card's type BEFORE the exile,
// because after it the card is in exile and "it WAS a creature card"
// is a question about the past.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     invasionOfInnistradOracleID + "#1",
		Name:         "Deluge of the Dead",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.CardID == source.InstanceID
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Deluge of the Dead — create two 2/2 Zombies",
					func(g *game.Game, item *game.StackItem) error {
						return CreateToken{
							Controller: item.Controller,
							Template:   BlackZombieToken(),
							N:          2,
						}.Apply(NewContext(g, item))
					})
			},
		}},
		Activated: []ActivatedAbility{{
			Label:   "{2}{B}: Exile target card from a graveyard. If it was a creature card, create a 2/2 black Zombie creature token.",
			Cost:    ManaCost("{2}{B}"),
			Targets: TargetCardInGraveyard("target card from a graveyard"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					c, ok := g.LookupCardForEffect(t.ID)
					if !ok {
						continue
					}
					wasCreature := c.IsCreature()
					if err := (ExileTarget{Target: t.ID}).Apply(ctx); err != nil {
						return err
					}
					if !wasCreature {
						continue
					}
					if err := (CreateToken{
						Controller: item.Controller,
						Template:   BlackZombieToken(),
						N:          1,
					}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			},
		}},
	})
}
