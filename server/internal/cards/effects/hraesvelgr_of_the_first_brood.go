package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hraesvelgr of the First Brood — Legendary Creature — Elder Dragon
// {4}{U}, 5/5:
//
//	"Flying, vigilance, ward {2}
//	 Shiva's Aid — When Hraesvelgr enters and whenever you cast a
//	 noncreature spell, target creature gets +1/+0 until end of turn and
//	 can't be blocked this turn."
//
// One ability with two triggers, so one row watching both events. The
// target is chosen as it goes on the stack and re-checked on resolution
// (CR 608.2b). The pump is a continuous effect on that object; "can't be
// blocked" is a rule, and both last the turn.
//
// No simplification.
func init() {
	shivasAid := OnAny([]game.EventKind{game.EventETB, game.EventCast},
		func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
			if ev.Kind == game.EventETB {
				return ev.CardID == source.InstanceID
			}
			return b10NoncreatureSpellCastByYou(ev, source, lki, g)
		},
		"Hraesvelgr of the First Brood — Shiva's Aid: target creature gets +1/+0 and can't be blocked this turn",
		func(g *game.Game, item *game.StackItem) error {
			ctx := NewContext(g, item)
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				if err := (BoostUntilEOT{Target: t.ID, Power: 1, Label: "Shiva's Aid — +1/+0"}).Apply(ctx); err != nil {
					return err
				}
				return RestrictUntilEOT{
					Target:       t.ID,
					Restrictions: game.CantBeBlocked,
					Label:        "Shiva's Aid — can't be blocked",
				}.Apply(ctx)
			}
			return nil
		})
	shivasAid.Targets = TargetCreature("target creature")
	Register(Spec{
		OracleID:        "44cbf70c-61cf-473e-901d-039956ced642",
		Name:            "Hraesvelgr of the First Brood",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
		Triggered: []game.TriggeredAbility{
			Ward(WardMana("{2}"), "Hraesvelgr of the First Brood — ward {2}"),
			shivasAid,
		},
	})
}
