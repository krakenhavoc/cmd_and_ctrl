package effects

import (
	"strconv"
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b846_helpers.go — shared pieces of Ice Age batch 846 (#5636, S82).

// OldB846SeraphNotePrefix prefixes the notes Seraph keeps of the creatures it
// brought back: "seraph:<instance id>:<object epoch>".
const OldB846SeraphNotePrefix = "seraph:"

// OldB846SeraphReturnBody is the delayed half of Seraph's first sentence: "put
// that card onto the battlefield under your control at the beginning of the
// next end step". The card is the graveyard object it became as it died
// (CR 400.7): one that has left the graveyard since stays where it went. It
// comes back under the delayed ability's controller, and, while Seraph is still
// there under that same control, Seraph notes it as one of its creatures, which
// "the creature" in the last sentence names. An on-disk key: never renamed,
// never reused.
var OldB846SeraphReturnBody = game.SimpleDelayedBody("delayed/old-b846-seraph-return",
	func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		for _, t := range item.Targets {
			if t.Kind != game.TargetCard || t.ID == uuid.Nil || !inGraveyardOneMoveAfter(g, t.ID, item.Params.Object.Epoch) {
				continue
			}
			if err := (ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneBattlefield, Controller: item.Controller}).Apply(ctx); err != nil {
				return err
			}
			back, ok := g.LookupCardForEffect(t.ID)
			if !ok || !OnBattlefield(g, t.ID) || back.Controller != item.Controller {
				continue
			}
			seraph, ok := g.LookupCardForEffect(item.SourceCardID)
			if !ok || SourceIsNewObject(g, item) || !OnBattlefield(g, seraph.InstanceID) || seraph.Controller != item.Controller {
				continue
			}
			g.AddNoteForEffect(seraph.InstanceID, OldB846SeraphNotePrefix+t.ID.String()+":"+strconv.Itoa(back.ObjectEpoch))
		}
		return nil
	})

// OldB846SeraphSacrificeTheCreatures is the last sentence: "Sacrifice the
// creature when you lose control of this creature". The notes are read as the
// ability resolves, off Seraph as it last existed if it has left (CR 608.2h);
// each creature is sacrificed only while it is still the object that came back
// (CR 400.7). A Seraph that is still on the battlefield, under new control,
// forgets them.
func OldB846SeraphSacrificeTheCreatures(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, n := range g.NotesWithPrefixForEffect(item.SourceCardID, OldB846SeraphNotePrefix) {
		idText, epochText, ok := strings.Cut(n, ":")
		if !ok {
			continue
		}
		id, err := uuid.Parse(idText)
		epoch, err2 := strconv.Atoi(epochText)
		if err != nil || err2 != nil {
			continue
		}
		c, found := g.LookupCardForEffect(id)
		if !found || c.ObjectEpoch != epoch || !OnBattlefield(g, id) {
			continue
		}
		if err := (SacrificePermanent{Target: id}).Apply(ctx); err != nil {
			return err
		}
	}
	g.ClearNotesWithPrefixForEffect(item.SourceCardID, OldB846SeraphNotePrefix)
	return nil
}
