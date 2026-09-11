package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Boros Charm — Instant {R}{W} (EDHREC rank 180):
//
//	"Choose one —
//	 • Boros Charm deals 4 damage to target player or planeswalker.
//	 • Permanents you control gain indestructible until end of turn.
//	 • Target creature gains double strike until end of turn."
//
// Three modes, two of them targeted — legal on a "choose one" (the
// per-mode target limit only bites when Max > 1), and the same shape
// Rakdos Charm already uses. The roadmap filed this card under
// "until end of turn"; #314 shipped the turn-scoped statics and the
// two duration modes are GrantKeywordUntilEOT.
//
// Mode 0's target clause is "player or planeswalker" — narrower than
// TargetAny (no creatures) and wider than TargetPlayer — so it is
// built by hand here rather than borrowed from targets.go.
//
// No simplifications as of S25 (#77). Mode 1 used to carry one:
// the indestructible grant appended a keyword string that nothing in
// the engine read, so the mode protected nothing and the mode label
// said so. S25 taught DestroyPermanentForEffect and the two
// damage-driven creature SBAs CR 702.12 (server/internal/game/
// indestructible.go), and — exactly as that note predicted — the
// mode started working without a line of card code changing. Only
// the disclaimer in the label came out.
//
// All three modes are now real: 4 damage, a turn-scoped
// indestructible grant over every permanent you control, and a
// double-strike grant the combat engine honours.
func init() {
	Register(Spec{
		OracleID: "2679d0dd-ba30-4a1c-b6a0-b3ac6c790496",
		Name:     "Boros Charm",
		Modes: ChooseOne(
			Mode("Boros Charm deals 4 damage to target player or planeswalker.",
				targetPlayerOrPlaneswalker("target player or planeswalker")),
			Mode("Permanents you control gain indestructible until end of turn."),
			Mode("Target creature gains double strike until end of turn.",
				TargetCreature("target creature")),
		),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			switch {
			case ctx.HasMode(0):
				if len(item.Targets) == 0 {
					return nil
				}
				return DealDamage{Source: ctx.Source(), Target: item.Targets[0].ID, Amount: 4}.Apply(ctx)
			case ctx.HasMode(1):
				return GrantKeywordUntilEOT{
					Match:    YouControl(),
					Keywords: []string{"indestructible"},
					Label:    "Boros Charm — permanents you control gain indestructible",
				}.Apply(ctx)
			case ctx.HasMode(2):
				if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
					return nil
				}
				return GrantKeywordUntilEOT{
					Target:   item.Targets[0].ID,
					Keywords: []string{"double strike"},
					Label:    "Boros Charm — target creature gains double strike",
				}.Apply(ctx)
			}
			return nil
		},
	})
}

// targetPlayerOrPlaneswalker is "target player or planeswalker" — a
// player, or a battlefield permanent that is a planeswalker. Kept
// file-local rather than added to targets.go so a concurrent batch
// touching that file doesn't collide; if a second card wants it,
// move it there.
func targetPlayerOrPlaneswalker(label string) *game.TargetSpec {
	return &game.TargetSpec{
		Mode:    "any",
		Label:   label,
		Players: true,
		Zones:   []game.ZoneKind{game.ZoneBattlefield},
		CardOK: func(_ *game.Game, _ uuid.UUID, c game.Card, _ game.ZoneKind) bool {
			return c.IsPlaneswalker()
		},
		Min: 1, Max: 1,
	}
}
