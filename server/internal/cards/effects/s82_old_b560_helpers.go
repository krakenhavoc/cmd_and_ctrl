package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b560_helpers.go — S82 set OLD, batch 560 (Coldsnap, pump and combat
// tricks).

// OldB560Crusader is the Spec of the two Coldsnap Crusaders (Stromgald
// Crusader, White Shield Crusader): "Protection from <color>; {S}: This
// creature gains flying until end of turn; {S}{S}: This creature gets +1/+0
// until end of turn", where {S} is the card's own colour symbol.
func OldB560Crusader(oracleID, name, symbol, protection string) Spec {
	one, two := "{"+symbol+"}", "{"+symbol+"}{"+symbol+"}"
	return Spec{
		OracleID:        oracleID,
		Name:            name,
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"protection from " + protection},
		Activated: []ActivatedAbility{
			S82OldB510ThisGainsRow(one, "flying"),
			ThisGetsRow(two, 1, 0, name),
		},
	}
}

// OldB560PumpPerOtherAttackingAurochs is "Whenever this creature attacks, it
// gets +1/+0 until end of turn for each other attacking Aurochs" (Bull
// Aurochs, Rimehorn Aurochs): the attacking Aurochs other than the source,
// counted as the trigger resolves, through the layered view.
func OldB560PumpPerOtherAttackingAurochs(label string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		n := 0
		// "Aurochs" is a word of the source's text (CR 612.2, #3984).
		aurochs := g.TextReaderForEffect(item).TextSubtype("Aurochs")
		for i := range g.Battlefield.Cards {
			c := &g.Battlefield.Cards[i]
			if c.InstanceID != item.SourceCardID && c.IsCreature() && c.HasSubtype(aurochs) && c.AttackingTarget != uuid.Nil {
				n++
			}
		}
		if n == 0 {
			return nil
		}
		return BoostUntilEOT{Target: item.SourceCardID, Power: n, Label: label}.Apply(NewContext(g, item))
	}
}

// OldB560TargetBlocksTarget is "Target creature blocks target creature this
// turn if able": the first target must block the second if it can (CR 509.1c),
// both checked as the ability resolves (CR 608.2b).
func OldB560TargetBlocksTarget(label string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		blocker, ok := ctx.ClauseTarget(0)
		if !ok || blocker.Kind != game.TargetCard {
			return nil
		}
		named, ok := ctx.ClauseTarget(1)
		if !ok || named.Kind != game.TargetCard {
			return nil
		}
		ref, ok := g.PermanentRefForEffect(named.ID)
		if !ok {
			return nil
		}
		return BlocksAttackerUntilEOT{Blocker: blocker.ID, Attacker: ref, Label: label}.Apply(ctx)
	}
}

// OldB560CardsNamedInEachGraveyard counts the cards named `name` in every
// player's graveyard (Sound the Call's Wolf). It walks the graveyards directly,
// since it runs inside a layer pass.
func OldB560CardsNamedInEachGraveyard(g *game.Game, name string) int {
	n := 0
	for _, p := range g.Seats {
		if p == nil || p.Graveyard == nil {
			continue
		}
		for _, c := range p.Graveyard.Cards {
			if c.Name == name {
				n++
			}
		}
	}
	return n
}

// PrintedOldB560SoundTheCallWolf is the 1/1 green Wolf creature token Sound
// the Call makes, with "This token gets +1/+1 for each card named Sound the
// Call in each graveyard." Its slug is an on-disk identity: never renamed,
// never reused.
func PrintedOldB560SoundTheCallWolf() TokenTemplate {
	return TokenTemplate{
		Slug: "wolf-sound-the-call",
		Card: game.Card{
			Name:      "Wolf",
			TypeLine:  "Token Creature — Wolf",
			Power:     1,
			Toughness: 1,
			Colors:    []string{"G"},
		},
		Static: []game.StaticAbility{{
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: SelfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, _ *game.Card) {
				n := OldB560CardsNamedInEachGraveyard(g, "Sound the Call")
				c.Power += n
				c.Toughness += n
			},
		}},
		Text: "This token gets +1/+1 for each card named Sound the Call in each graveyard.",
	}
}

// OldB560ExileInsteadLinkedToSource is the Replace of "If another creature
// would die, exile it instead" for a source whose later text reads the cards it
// exiled (Void Maw): the move goes to exile and the card is "exiled with" the
// source object (CR 607.2a), stamped once it has really landed.
func OldB560ExileInsteadLinkedToSource(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) error {
	ev.NewZone, ev.NewZoneOwner = game.ZoneExile, uuid.Nil
	ev.ExiledWith = game.PermissionCardRef{ID: src.InstanceID, Epoch: src.ObjectEpoch}
	return nil
}
