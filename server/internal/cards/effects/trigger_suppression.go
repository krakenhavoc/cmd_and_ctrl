package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// trigger_suppression.go: the vocabulary for "[an event] doesn't cause
// abilities to trigger" (#1735). This file pairs with trigger_doubling.go.
// The engine side, including which board a suppressor is read from, is
// game/trigger_suppression.go.
//
//	Torpor Orb, Hushwing Gryff,   SuppressesEntering(Creature())
//	Tocatli Honor Guard
//	Doorkeeper Thrull             SuppressesEntering(Or(Artifact(), Creature()))
//	Hushbringer                   SuppressesEntering(Creature()) +
//	                              SuppressesDying(Creature())
//	Elesh Norn, Mother of         OfOpponentsPermanents(SuppressesEntering(nil))
//	Machines (third paragraph)
//
// The cause helpers use the doubler's own event predicates (isEntering,
// isDying), so "a permanent entering" means the same thing on both
// halves of Elesh Norn.

// SuppressesEntering declares "[filter] entering don't cause abilities
// to trigger". The filter is judged against the entering permanent as
// it is on the battlefield (CR 603.6a). A nil filter means any
// permanent. That covers both the entering permanent's own "when this
// enters" and every other "whenever a [thing] enters", and it covers
// a token being created (CR 701.7a puts it onto the battlefield).
func SuppressesEntering(filter CardPredicate) game.TriggerSuppressor {
	return game.TriggerSuppressor{
		Label: "entering",
		Suppresses: func(g *game.Game, q game.TriggerSuppressionQuery) bool {
			if !q.HasSubject || !isEntering(q.Event) {
				return false
			}
			return filter == nil || filter(g, q.SuppressorLKI.Controller, suppressionSubject(g, q))
		},
	}
}

// SuppressesDying declares "[filter] dying don't cause abilities to
// trigger" (Hushbringer). The subject is the permanent's last-known
// battlefield object, as it is for DoublesDying. A "leaves the
// battlefield" ability is stopped when the permanent died, because
// the death is the event that triggered it, and the doubler counts it
// the same way.
func SuppressesDying(filter CardPredicate) game.TriggerSuppressor {
	return game.TriggerSuppressor{
		Label: "dying",
		Suppresses: func(g *game.Game, q game.TriggerSuppressionQuery) bool {
			if !q.HasSubject || !isDying(q.Event) {
				return false
			}
			return filter == nil || filter(g, q.SuppressorLKI.Controller, suppressionSubject(g, q))
		},
	}
}

// OfOpponentsPermanents narrows a suppressor to "abilities of
// permanents your opponents control" (Elesh Norn, Mother of Machines).
// It does not narrow to a spell's "when you cast" ability, an emblem's
// ability or a graveyard card's ability, because none of those belongs
// to a permanent. Your own permanents' abilities are never narrowed
// in.
func OfOpponentsPermanents(s game.TriggerSuppressor) game.TriggerSuppressor {
	inner := s.Suppresses
	s.Suppresses = func(g *game.Game, q game.TriggerSuppressionQuery) bool {
		if !q.SourceIsPermanent || q.SourceLKI.Controller == q.SuppressorLKI.Controller {
			return false
		}
		return inner != nil && inner(g, q)
	}
	return s
}

// suppressionSubject is subjectCard for a suppression query.
func suppressionSubject(g *game.Game, q game.TriggerSuppressionQuery) game.Card {
	return eventSubjectCard(g, q.Subject, q.SubjectLKI)
}

// CreaturesEnteringDontTrigger is the printed line "Creatures entering
// don't cause abilities to trigger", labelled with the card's name.
// Torpor Orb, Hushwing Gryff, Tocatli Honor Guard and the first half of
// Hushbringer print it word for word.
func CreaturesEnteringDontTrigger(label string) game.TriggerSuppressor {
	s := SuppressesEntering(Creature())
	s.Label = label
	return s
}
