package game

import (
	"github.com/google/uuid"
)

// attach_transform.go — "as this permanent transforms into <Aura face>,
// attach it to a player" (Curse of Leeches // Leeching Lurker, #2586;
// ADR 0079's amendment of 2026-10-09).
//
// # The shape
//
// The permanent is already on the battlefield, so this is not the cast
// path (the Aura spell's "enchant player" is a TARGET, announced and
// re-checked) and not the CR 303.4g "enters as an Aura" path. It is an
// instruction the face itself prints, run by the transform verb through
// CardDef.AsTransformsInto, and the player is CHOSEN by the Aura's
// controller as the face turns over: any player the enchant clause
// admits, themselves included. Nothing targets, so hexproof is
// irrelevant to it; protection is not, because CR 702.16 says a player
// with protection from the Aura's quality can't be enchanted by it, and
// the prompt does not offer such a seat.
//
// The prompt is the existing option_pick over seats (choose_player.go),
// so the choice gate, the enumerator, the wire, the bot's first-offer
// policy and the CR 800.4a pruning all carry it unchanged. What is new
// is the one thing an ordinary choose-a-player never needed: the card
// is an Aura sitting on the battlefield attached to NOTHING while the
// question is open, and CR 704.5m would put it in the graveyard at the
// next state-based check. attachPromptOpenForLocked is the exemption —
// it reads the open prompt itself rather than a flag on the card, so
// there is no state to clone, snapshot or forget to clear.
//
// # When there is no one to attach it to
//
// Every seat refused (all eliminated, or each protected from the Aura),
// or the chooser gone: nothing is queued, the Aura is unattached with no
// open question, and CR 704.5m sends it to its owner's graveyard. That is
// what the rules say happens to an Aura "not attached to an object or
// player", and it needs no card code. A prompt dropped unanswered
// (#1006) ends the same way.

// QueueAttachSourceToPlayerForEffect asks the Aura `source`'s
// controller which player it is to enchant and, on the answer, attaches
// it. Returns the choice ID, or uuid.Nil when nothing was queued (the
// permanent is not an Aura on the battlefield, or no player may be
// enchanted).
//
// The candidates are the seats the Aura's own enchant clause admits at
// this moment, exactly the predicate CR 704.5m will hold it to
// afterwards (attachmentLegalLocked): the catalogued TargetSpec without
// the targeting gate, then CR 702.16's protection test against the
// Aura as a source. Ordered by the shared rule (most life first), which
// is the bot's policy for a prompt it has nothing to say about.
//
// Caller must hold g.mu (an AsTransformsInto hook does).
func (g *Game) QueueAttachSourceToPlayerForEffect(source uuid.UUID, question string) uuid.UUID {
	c := findBattlefieldCard(g, source)
	if c == nil || !c.IsAura() {
		return uuid.Nil
	}
	spec := TargetSpecFor(catalogKeyOf(c))
	var among []uuid.UUID
	for _, p := range g.Seats {
		if p == nil || p.Eliminated {
			continue
		}
		if spec != nil && !g.specMatchLocked(SourceChooser(c.Controller), spec, TargetRef{Kind: TargetPlayer, ID: p.ID}, false) {
			continue
		}
		if _, refused := g.PlayerProtectedFromLocked(p, SourceCharacteristics(c)); refused {
			continue
		}
		among = append(among, p.ID)
	}
	eligible := g.eligibleChosenPlayersLocked(among)
	if len(eligible) == 0 {
		return uuid.Nil
	}
	return g.QueueOptionPickForEffect(OptionPickPrompt{
		Chooser:  c.Controller,
		Source:   source,
		Question: question,
		Options:  g.seatChoiceOptionsLocked(eligible),
		// Keyed (#4361 slice 6). The key is also the marker
		// attachPromptOpenForLocked reads, so it survives a restore.
		ThenKey: attachSourceToPlayerThen,
		Carry:   []uuid.UUID{source},
	})
}

// attachSourceToPlayerThen attaches the Aura, Carry's one ID, to the
// seat chosen, read off the option rather than an index for #994's
// reason: a prune that renumbers the list cannot attach the Aura to the
// player one place along. A dropped prompt (#1006) chooses nobody: the
// Aura stays unattached and CR 704.5m takes it. An on-disk key: never
// renamed, never reused.
var attachSourceToPlayerThen OptionPickThen

// Registered from init: the continuation can reach code that names the
// key, which a package-level initializer cannot.
func init() {
	attachSourceToPlayerThen = RegisterOptionPickThen("option-pick/attach-source-to-player", attachSourceToPlayerAnswered)
}

func attachSourceToPlayerAnswered(g *Game, r OptionPicked) error {
	chosen := chosenSeat(r)
	if len(r.Carry) == 0 || chosen == uuid.Nil {
		return nil
	}
	return g.AttachForEffect(r.Carry[0], TargetRef{Kind: TargetPlayer, ID: chosen})
}

// attachPromptOpenForLocked reports whether the Aura `id` has its
// "attach it to a player" question open, in which case it is
// legitimately attached to nothing for now and CR 704.5m must wait.
//
// Caller must hold g.mu.
func (g *Game) attachPromptOpenForLocked(id uuid.UUID) bool {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == PendingChoiceOptionPick && c.Source == id &&
			c.PickThen == attachSourceToPlayerThen.key {
			return true
		}
	}
	return false
}
