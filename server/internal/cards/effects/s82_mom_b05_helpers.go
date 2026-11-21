package effects

import (
	"strconv"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_mom_b05_helpers.go — shared shapes for March of the Machine, S82
// batch 5 (#5477).

// BackupAlso is Backup (soc_b6_helpers.go) for a creature whose printed
// abilities below the keyword include a rule that is a layer effect
// rather than a keyword or an ability bundle: Chomping Kavu's "can't be
// blocked by creatures with power 2 or less" and Streetwise Negotiator's
// "assigns combat damage equal to its toughness rather than its power".
// `mods` are those effects, applied in the same until-end-of-turn
// effect as the keywords and the bundle, and only to another creature
// (CR 702.165a).

func BackupAlso(name string, n int, grantKey string, mods []game.Mod, keywords ...string) game.TriggeredAbility {
	kws := append([]string(nil), keywords...)
	t := Targeting(WhenThisEnters(name+" — backup "+strconv.Itoa(n), func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		for _, t := range ctx.LegalTargets() {
			if t.Kind != game.TargetCard {
				continue
			}
			if err := (AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: n}).Apply(ctx); err != nil {
				return err
			}
			if t.ID == item.SourceCardID || !OnBattlefield(g, t.ID) {
				return nil
			}
			also := append([]game.Mod(nil), mods...)
			if len(kws) > 0 {
				also = append(also, game.AddKeywordsMod(kws...))
			}
			if grantKey == "" {
				if len(also) == 0 {
					return nil
				}
				return ScopedEffectFor{Target: t.ID, Mods: also, Duration: DurationUntilEndOfTurn(ctx),
					Label: name + " — backup: the creature gains its abilities until end of turn"}.Apply(ctx)
			}
			return GrantAbilitiesFor{Target: t.ID, Keys: []string{grantKey}, Also: also,
				Label: name + " — backup: the creature gains its abilities until end of turn"}.Apply(ctx)
		}
		return nil
	}), TargetCreature("target creature"))
	t.Keyword = game.KeywordBackup
	return t
}

// momB05PlusOneCountersFirstTimeThisTurn is the condition behind Axgard
// Artisan and Botanical Brawler: the event is a placement that raised
// the +1/+1 count of the object it names, and it is the first one this
// turn (Game.TimesPlusOneCountersPutThisTurn, written ahead of the
// trigger harvester). The object is the event's target; the caller
// judges whose permanent it is.
func MomB05PlusOneCountersFirstTimeThisTurn(ev game.Event, g *game.Game) bool {
	if ev.Kind != game.EventCounterPlaced || ev.Target == uuid.Nil ||
		B33CountersPlacedDelta(ev, game.CounterPlusOne, g) <= 0 {
		return false
	}
	return g.TimesPlusOneCountersPutThisTurn(ev.Target) == 1
}
