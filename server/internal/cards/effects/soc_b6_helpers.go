package effects

import (
	"strconv"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// soc_b6_helpers.go — shared shapes for Secrets of Strixhaven Commander,
// batch 6 (#4834).

// Backup is CR 702.165a's keyword ability:
//
//	"Backup N (When this creature enters, put N +1/+1 counters on target
//	 creature. If that's another creature, it also gains the non-backup
//	 abilities of this creature printed below this one until end of
//	 turn.)"
//
// `keywords` are the keyword abilities printed below backup and
// `grantKey` names the Spec.Grants bundle that holds the rest of them
// (empty when there are none). Both are the card's printed text, fixed
// in the catalog, so what the target gains is decided as the ability is
// put on the stack and an ability the creature has gained or lost since
// does not change it (CR 702.165c, 702.165d). The target is any creature,
// this one included; the counters go on whichever was chosen and is still
// legal as the ability resolves (CR 608.2b), and the abilities only on
// another creature, in one until-end-of-turn effect.
//
// The row carries game.KeywordBackup (#5651), stamped here and nowhere
// else, so its item on the stack is a backup ability to a watcher of
// "becomes the target of a backup ability" (Event.TargeterKeyword).
func Backup(name string, n int, grantKey string, keywords ...string) game.TriggeredAbility {
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
			if t.ID == item.SourceCardID || !onBattlefield(g, t.ID) {
				return nil
			}
			var also []game.Mod
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

// socB6AurasOnTheDead is the Auras that were attached to the creature an
// EventLTB names as it died, each with the player who controlled it then
// (CR 603.10a, 608.2h): the creature's last-known attachments, plus any
// Aura that left the battlefield in the same event batch while attached to
// it, which a simultaneous destruction can record a moment before the
// creature itself (Hateful Eidolon's ruling). Each Aura is read as the
// object it was: live while it is still that object on the battlefield,
// from its last-known record once it has left.
func socB6AurasOnTheDead(g *game.Game, ev game.Event) map[uuid.UUID]uuid.UUID {
	info, ok := g.LastKnownPermanentForEffect(ev.CardID)
	if !ok {
		return nil
	}
	out := map[uuid.UUID]uuid.UUID{}
	for _, ref := range info.Attachments {
		if c, live := g.LookupCardForEffect(ref.ID); live && c.ObjectEpoch == ref.Epoch && onBattlefield(g, ref.ID) {
			if c.HasSubtype("Aura") {
				out[ref.ID] = c.Controller
			}
			continue
		}
		if rec, found := g.LastKnownPermanentForEffect(ref.ID); found && rec.Epoch == ref.Epoch &&
			keywordSliceContains(rec.Characteristic.Subtypes, "Aura") {
			out[ref.ID] = rec.Controller
		}
	}
	if ev.Batch == 0 {
		return out
	}
	for _, e := range g.EventsThisTurn() {
		if e.Kind != game.EventLTB || e.Batch != ev.Batch || e.CardID == ev.CardID {
			continue
		}
		if _, seen := out[e.CardID]; seen {
			continue
		}
		rec, found := g.LastKnownPermanentForEffect(e.CardID)
		if !found || rec.AttachedTo.Kind != game.TargetCard || rec.AttachedTo.ID != ev.CardID ||
			(rec.AttachedObject != nil && rec.AttachedObject.Epoch != info.Epoch) ||
			!keywordSliceContains(rec.Characteristic.Subtypes, "Aura") {
			continue
		}
		out[e.CardID] = rec.Controller
	}
	return out
}
