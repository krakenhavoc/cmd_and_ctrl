package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tree of Redemption — Creature — Plant {3}{G}, 0/13 (Tarkir: Dragonstorm
// Commander):
//
//	"Defender
//	 {T}: Exchange your life total with this creature's toughness."
//
// CR 701.12g: an exchange of two numerical values makes each become the other's
// previous value. The toughness read is the Tree's current one (modifiers
// included, the ruling) and your life total is read at the same moment. Your
// life becomes the old toughness by the necessary gain or loss, which goes
// through the ordinary life-change pipeline (replacement effects may modify it,
// triggers may trigger on it, and a player who can't gain or lose life can't be
// moved that way). The Tree's toughness is set to the old life total by a
// continuous effect with no stated duration (CR 611.2a, 613.4b): a layer 7b
// base-toughness effect pinned to this object that lasts as long as it is the
// same permanent, with Auras, Equipment and counters applying after it (the
// ruling), and a newer exchange overriding an older by timestamp. A Tree that
// has left the battlefield, or is a new object, does nothing (the ruling).
// Toughness 0 or less puts it into the graveyard as a state-based action
// (CR 704.5f).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "1372dd15-cdb9-470c-b3e8-3423697e46e8",
		Name:            "Tree of Redemption",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Activated: []ActivatedAbility{{
			Label: "{T}: Exchange your life total with this creature's toughness.",
			Cost:  TapCost(),
			// It can raise the Tree's toughness to your life total (ADR 0142).
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Effect: func(g *game.Game, item *game.StackItem) error {
				if g.AbilitySourceGoneForEffect(item) {
					return nil
				}
				ctx := NewContext(g, item)
				p := g.PlayerByIDForEffect(item.Controller)
				if p == nil || !onBattlefield(g, item.SourceCardID) {
					return nil
				}
				g.RecomputeLayersIfStaleLocked()
				tree, ok := g.LookupCardForEffect(item.SourceCardID)
				if !ok {
					return nil
				}
				toughness, life := tree.CurrentToughness(), p.Life
				if err := (ScopedEffectFor{
					Target:   item.SourceCardID,
					Mods:     []game.Mod{game.SetBaseToughnessMod(life)},
					Duration: g.PinnedTo(game.IndefiniteDuration(), item.SourceCardID),
					Label:    "Tree of Redemption — toughness becomes your former life total",
				}).Apply(ctx); err != nil {
					return err
				}
				if delta := toughness - life; delta != 0 {
					return g.ChangePlayerLifeForEffect(item.SourceCardID, item.Controller, delta)
				}
				return nil
			},
		}},
	})
}
