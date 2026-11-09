package lobby

// pods_http.go is ADR 0220 slice 2 (#4880): pods, over internal/pods.
//
//	GET    /me/pods                             my pods, the invites waiting on me (suggestions: slice 4)
//	POST   /pods                                make a pod: {name}
//	GET    /pods/{id}                           members only: the pod
//	PATCH  /pods/{id}                           owner: rename, {name}
//	DELETE /pods/{id}                           owner: delete
//	PUT    /pods/{id}/owner                     owner: hand over, {user_id}
//	POST   /pods/{id}/invites                   members: invite a friend, {user_id}
//	DELETE /pods/{id}/invites/{user_id}         the inviter or the owner: cancel
//	POST   /me/pod-invites/{pod_id}/accept      the invitee
//	POST   /me/pod-invites/{pod_id}/decline     the invitee, silently
//	POST   /pods/{id}/link                      owner: mint or rotate, answers the token once
//	DELETE /pods/{id}/link                      owner: turn off
//	GET    /pods/join/{token}                   anyone signed in: the preview
//	POST   /pods/join/{token}                   anyone signed in: join
//	DELETE /pods/{id}/members/{user_id}         owner: remove; `me` or your own id: leave
//	PUT    /admin/pods/{id}/name                admin: rename, {name}
//	DELETE /admin/pods/{id}                     admin: delete
//
// The caller is a signed-in person. A guest seat, a guest spectator and
// the admin token are credentials, not people, and get 403 ("pods need
// an account"). With no database (c.Pods nil) every route answers 503.
//
// A PRIVATE POD (§1): every route answers 404 to anyone not in the pod,
// for a pod that exists and one that does not alike.
//
// ONE PATTERN FOR TWO SHAPES: /pods/join/{token} and /pods/{id}/link
// both match /pods/join/link, so Go's ServeMux refuses to hold both as
// patterns. GET and POST /pods/{id}/{action} are registered instead and
// podAction dispatches: id "join" is the link, anything else a pod. A
// pod id is a UUID, so it is never "join". Slice 5's GET
// /pods/{id}/stats and /leaderboard join the GET dispatcher.
//
// REFUSALS THAT DO NOT TELL (§3b, §6): an unknown, rotated, expired or
// turned-off link, a full pod and a join refused for a block all get
// one answer. An invite refused because the friend has a block with a
// member does not say who.

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/auth"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/moderation"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/pods"
)

// The words of the refusals (ADR 0220 §3, §5).
const (
	errPodsOff           = "pods are not available on this server"
	errPodsNeedAccount   = "pods need an account; sign in with Discord to make or join one"
	errNoSuchPod         = "no such pod"
	errNotPodOwner       = "only the pod's owner can do that"
	errPodNameInvalid    = "a pod's name is 1 to 40 characters on one line"
	errTooManyPods       = "You're in 10 pods. Leave one to join another."
	errTooManyNewPods    = "you've made a lot of pods lately; try again later"
	errPodFull           = "This pod is full"
	errDeadPodLink       = "This invite link can't be used."
	errNotAFriend        = "no such friend; you can invite one of your friends"
	errAlreadyInPod      = "that person is already in this pod"
	errPodInviteRefused  = "can't invite that person to this pod"
	errTooManyPodInvites = "this pod has 20 invites waiting for an answer; cancel one first"
	errPodInvitesHour    = "you've sent a lot of pod invites in the last hour; try again later"
	errNoPodInvite       = "no such pod invite"
	errNotInviter        = "only the person who sent an invite, or the pod's owner, can cancel it"
	errNoPodMember       = "no such member"
	errNoSuchSuggestion  = "no such suggestion"
)

// PodPerson is someone a pod shows.
type PodPerson struct {
	UserID      string `json:"user_id,omitempty"`
	DisplayName string `json:"display_name"`
	AvatarURL   string `json:"avatar_url,omitempty"`
}

// podSummary is one pod in GET /me/pods, and what making, accepting and
// joining answer. MemberCount leaves out a member with a block with the
// caller (§6).
type podSummary struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	MemberCount int    `json:"member_count"`
	Role        string `json:"role"`
}

// podInviteIn is an invite waiting on the caller. Since is when it was
// sent, Unix milliseconds.
type podInviteIn struct {
	PodID     string    `json:"pod_id"`
	PodName   string    `json:"pod_name"`
	InvitedBy PodPerson `json:"invited_by"`
	Since     int64     `json:"since"`
}

// myPodsResponse is GET /me/pods. Suggestions is slice 4's and empty
// until then.
type myPodsResponse struct {
	Pods        []podSummary  `json:"pods"`
	Invites     []podInviteIn `json:"invites"`
	Suggestions []struct{}    `json:"suggestions"`
	MaxPods     int           `json:"max_pods"`
	MaxMembers  int           `json:"max_members"`
}

// podMember is one row of GET /pods/{id}'s members. A row with Blocked
// is the one an owner keeps for a member they blocked: its id and name,
// for Remove, and nothing else, so JoinedAt and Sharing are absent.
type podMember struct {
	PodPerson
	JoinedAt *int64 `json:"joined_at,omitempty"`
	Sharing  *bool  `json:"sharing,omitempty"`
	Owner    bool   `json:"owner,omitempty"`
	Blocked  bool   `json:"blocked,omitempty"`
}

// podPendingInvite is one invite the pod waits on. The invitee's user id
// and CanCancel are there only for the inviter and the owner, who can
// cancel it. Mine marks the caller's own.
type podPendingInvite struct {
	PodPerson
	InvitedBy PodPerson `json:"invited_by"`
	Since     int64     `json:"since"`
	Mine      bool      `json:"mine,omitempty"`
	CanCancel bool      `json:"can_cancel,omitempty"`
}

// podLinkState is the owner's view of the invite link: whether one is
// live and when it stops working.
type podLinkState struct {
	Live      bool  `json:"live"`
	ExpiresAt int64 `json:"expires_at,omitempty"`
}

// podResponse is GET /pods/{id}.
type podResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Role string `json:"role"`
	// Owner is absent when the owner and the caller have a block between
	// them (§6).
	Owner       *PodPerson         `json:"owner,omitempty"`
	MemberCount int                `json:"member_count"`
	MaxMembers  int                `json:"max_members"`
	Members     []podMember        `json:"members"`
	Invites     []podPendingInvite `json:"invites"`
	// Link is the owner's only.
	Link *podLinkState `json:"link,omitempty"`
}

// podLinkResponse is a freshly made invite link, shown this once.
type podLinkResponse struct {
	Token     string `json:"token"`
	Path      string `json:"path"`
	ExpiresAt int64  `json:"expires_at"`
}

// podPreviewResponse is GET /pods/join/{token}. PodID is there only for
// someone already in the pod.
type podPreviewResponse struct {
	PodID       string `json:"pod_id,omitempty"`
	Name        string `json:"name"`
	OwnerName   string `json:"owner_name"`
	MemberCount int    `json:"member_count"`
	Member      bool   `json:"member"`
}

// podInviteResponse is POST /pods/{id}/invites: "sent", or "pending" when
// one already was.
type podInviteResponse struct {
	Outcome     string `json:"outcome"`
	UserID      string `json:"user_id"`
	DisplayName string `json:"display_name,omitempty"`
}

type podNameBody struct {
	Name string `json:"name"`
}

type newPodBody struct {
	Name       string `json:"name"`
	Suggestion string `json:"suggestion,omitempty"`
}

type podUserBody struct {
	UserID string `json:"user_id"`
}

// podCaller is the signed-in person behind a pod route, once the feature
// is known to be on.
func podCaller(c Config, r *http.Request) (string, error) {
	p, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		return "", httpError(http.StatusInternalServerError, "missing principal")
	}
	if !isSignedInPerson(p) {
		return "", httpError(http.StatusForbidden, errPodsNeedAccount)
	}
	if c.Pods == nil {
		return "", httpError(http.StatusServiceUnavailable, errPodsOff)
	}
	return p.UserID.String(), nil
}

// podUserID is a path or body user id as a uuid string.
func podUserID(raw string) (string, error) {
	id, err := uuid.Parse(strings.TrimSpace(raw))
	if err != nil || id == uuid.Nil {
		return "", httpError(http.StatusBadRequest, errInvalidUserID)
	}
	return id.String(), nil
}

// podError maps a pods error onto the route's answer.
func podError(c Config, err error, what string) error {
	switch {
	case errors.Is(err, moderation.ErrNameNotAllowed):
		// writeLobbyError answers it 400 with code name_not_allowed.
		return err
	case errors.Is(err, pods.ErrNotFound):
		return httpError(http.StatusNotFound, errNoSuchPod)
	case errors.Is(err, pods.ErrNotOwner):
		return httpError(http.StatusForbidden, errNotPodOwner)
	case errors.Is(err, pods.ErrNameInvalid):
		return httpError(http.StatusBadRequest, errPodNameInvalid)
	case errors.Is(err, pods.ErrTooManyPods):
		return httpError(http.StatusConflict, errTooManyPods)
	case errors.Is(err, pods.ErrCreateRateLimited):
		return httpError(http.StatusTooManyRequests, errTooManyNewPods)
	case errors.Is(err, pods.ErrPodFull):
		return httpError(http.StatusConflict, errPodFull)
	case errors.Is(err, pods.ErrDeadLink):
		return httpError(http.StatusNotFound, errDeadPodLink)
	case errors.Is(err, pods.ErrNotFriend):
		return httpError(http.StatusNotFound, errNotAFriend)
	case errors.Is(err, pods.ErrAlreadyMember):
		return httpError(http.StatusConflict, errAlreadyInPod)
	case errors.Is(err, pods.ErrInviteRefused):
		return httpError(http.StatusForbidden, errPodInviteRefused)
	case errors.Is(err, pods.ErrTooManyInvites):
		return httpError(http.StatusConflict, errTooManyPodInvites)
	case errors.Is(err, pods.ErrInviteRateLimited):
		return httpError(http.StatusTooManyRequests, errPodInvitesHour)
	case errors.Is(err, pods.ErrNoInvite):
		return httpError(http.StatusNotFound, errNoPodInvite)
	case errors.Is(err, pods.ErrNotInviter):
		return httpError(http.StatusForbidden, errNotInviter)
	case errors.Is(err, pods.ErrNoMember):
		return httpError(http.StatusNotFound, errNoPodMember)
	case errors.Is(err, pods.ErrSelf):
		return httpError(http.StatusBadRequest, errCannotTargetSelf)
	}
	c.logger().Error(what+" failed", "err", err)
	return httpError(http.StatusInternalServerError, "something went wrong; try again")
}

func summaryOf(s pods.Summary) podSummary {
	return podSummary{ID: s.ID, Name: s.Name, MemberCount: s.Members, Role: string(s.Role)}
}

func linkResponse(token string, exp time.Time) podLinkResponse {
	return podLinkResponse{Token: token, Path: "#/pods/join/" + token, ExpiresAt: exp.UnixMilli()}
}

// myPods is GET /me/pods.
func myPods(c Config, w http.ResponseWriter, r *http.Request) error {
	me, err := podCaller(c, r)
	if err != nil {
		return err
	}
	o, err := c.Pods.Mine(r.Context(), me)
	if err != nil {
		return podError(c, err, "listing a user's pods")
	}
	out := myPodsResponse{
		Pods: make([]podSummary, 0, len(o.Pods)), Invites: make([]podInviteIn, 0, len(o.Invites)),
		Suggestions: []struct{}{}, MaxPods: pods.MaxPodsPerPerson, MaxMembers: pods.MaxMembers,
	}
	for _, p := range o.Pods {
		out.Pods = append(out.Pods, summaryOf(p))
	}
	for _, in := range o.Invites {
		out.Invites = append(out.Invites, podInviteIn{
			PodID: in.PodID, PodName: in.PodName, Since: in.At.UnixMilli(),
			InvitedBy: PodPerson{DisplayName: in.From.DisplayName, AvatarURL: in.From.AvatarURL},
		})
	}
	w.Header().Set("Cache-Control", "no-store")
	return writeJSON(w, http.StatusOK, out)
}

// postPod is POST /pods. A suggestion is slice 4's; until it lands, no
// group key names one.
func postPod(c Config, w http.ResponseWriter, r *http.Request) error {
	me, err := podCaller(c, r)
	if err != nil {
		return err
	}
	var body newPodBody
	if err := decodeJSON(w, r, &body); err != nil {
		return err
	}
	if strings.TrimSpace(body.Suggestion) != "" {
		return httpError(http.StatusNotFound, errNoSuchSuggestion)
	}
	s, err := c.Pods.Create(r.Context(), me, body.Name)
	if err != nil {
		return podError(c, err, "making a pod")
	}
	return writeJSON(w, http.StatusCreated, summaryOf(s))
}

// getPod is GET /pods/{id}.
func getPod(c Config, w http.ResponseWriter, r *http.Request) error {
	me, err := podCaller(c, r)
	if err != nil {
		return err
	}
	d, err := c.Pods.Get(r.Context(), me, r.PathValue("id"))
	if err != nil {
		return podError(c, err, "reading a pod")
	}
	owner := d.Role == pods.RoleOwner
	out := podResponse{
		ID: d.ID, Name: d.Name, Role: string(d.Role), MaxMembers: pods.MaxMembers,
		Members: make([]podMember, 0, len(d.Members)), Invites: make([]podPendingInvite, 0, len(d.Invites)),
	}
	if d.Owner != nil {
		out.Owner = &PodPerson{UserID: d.Owner.UserID, DisplayName: d.Owner.DisplayName, AvatarURL: d.Owner.AvatarURL}
	}
	for _, m := range d.Members {
		row := podMember{PodPerson: PodPerson{UserID: m.UserID, DisplayName: m.DisplayName}}
		if m.Blocked {
			row.Blocked = true
		} else {
			joined, sharing := m.JoinedAt.UnixMilli(), m.Sharing
			row.AvatarURL, row.JoinedAt, row.Sharing, row.Owner = m.AvatarURL, &joined, &sharing, m.Owner
		}
		out.Members = append(out.Members, row)
	}
	out.MemberCount = len(out.Members)
	for _, in := range d.Invites {
		mine := in.InvitedBy.UserID == me
		row := podPendingInvite{
			PodPerson: PodPerson{DisplayName: in.To.DisplayName, AvatarURL: in.To.AvatarURL},
			InvitedBy: PodPerson{DisplayName: in.InvitedBy.DisplayName, AvatarURL: in.InvitedBy.AvatarURL},
			Since:     in.At.UnixMilli(), Mine: mine, CanCancel: mine || owner,
		}
		if row.CanCancel {
			row.UserID = in.To.UserID
		}
		out.Invites = append(out.Invites, row)
	}
	if owner {
		out.Link = &podLinkState{}
		if !d.LinkExpiresAt.IsZero() {
			out.Link.Live, out.Link.ExpiresAt = true, d.LinkExpiresAt.UnixMilli()
		}
	}
	w.Header().Set("Cache-Control", "no-store")
	return writeJSON(w, http.StatusOK, out)
}

// patchPod is PATCH /pods/{id}: rename.
func patchPod(c Config, w http.ResponseWriter, r *http.Request) error {
	me, err := podCaller(c, r)
	if err != nil {
		return err
	}
	var body podNameBody
	if err := decodeJSON(w, r, &body); err != nil {
		return err
	}
	id := r.PathValue("id")
	name, err := c.Pods.Rename(r.Context(), me, id, body.Name)
	if err != nil {
		return podError(c, err, "renaming a pod")
	}
	return writeJSON(w, http.StatusOK, map[string]string{"id": id, "name": name})
}

// deletePod is DELETE /pods/{id}.
func deletePod(c Config, w http.ResponseWriter, r *http.Request) error {
	me, err := podCaller(c, r)
	if err != nil {
		return err
	}
	if err := c.Pods.Delete(r.Context(), me, r.PathValue("id")); err != nil {
		return podError(c, err, "deleting a pod")
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}

// putPodOwner is PUT /pods/{id}/owner: hand the pod to another member.
func putPodOwner(c Config, w http.ResponseWriter, r *http.Request) error {
	me, err := podCaller(c, r)
	if err != nil {
		return err
	}
	var body podUserBody
	if err := decodeJSON(w, r, &body); err != nil {
		return err
	}
	to, err := podUserID(body.UserID)
	if err != nil {
		return err
	}
	if err := c.Pods.HandOver(r.Context(), me, r.PathValue("id"), to); err != nil {
		return podError(c, err, "handing a pod over")
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}

// podAction is POST /pods/{id}/{action}: /pods/{id}/invites,
// /pods/{id}/link, and /pods/join/{token} (see the file comment).
func podAction(c Config, w http.ResponseWriter, r *http.Request) error {
	if r.PathValue("id") == "join" {
		return joinPodByLink(c, w, r, r.PathValue("action"))
	}
	switch r.PathValue("action") {
	case "invites":
		return postPodInvite(c, w, r)
	case "link":
		return postPodLink(c, w, r)
	}
	if _, err := podCaller(c, r); err != nil {
		return err
	}
	return httpError(http.StatusNotFound, "not found")
}

// podRead is GET /pods/{id}/{action}: /pods/join/{token} now, and slice
// 5's stats routes later.
func podRead(c Config, w http.ResponseWriter, r *http.Request) error {
	if r.PathValue("id") == "join" {
		return previewPodLink(c, w, r, r.PathValue("action"))
	}
	if _, err := podCaller(c, r); err != nil {
		return err
	}
	return httpError(http.StatusNotFound, "not found")
}

// postPodInvite is POST /pods/{id}/invites: invite one of your friends.
func postPodInvite(c Config, w http.ResponseWriter, r *http.Request) error {
	me, err := podCaller(c, r)
	if err != nil {
		return err
	}
	var body podUserBody
	if err := decodeJSON(w, r, &body); err != nil {
		return err
	}
	to, err := podUserID(body.UserID)
	if err != nil {
		return err
	}
	outcome, err := c.Pods.Invite(r.Context(), me, r.PathValue("id"), to)
	if err != nil {
		return podError(c, err, "sending a pod invite")
	}
	out := podInviteResponse{Outcome: string(outcome), UserID: to}
	if id, perr := uuid.Parse(to); perr == nil {
		out.DisplayName = c.displayNameOf(r.Context(), id)
	}
	return writeJSON(w, http.StatusOK, out)
}

// deletePodInvite is DELETE /pods/{id}/invites/{user_id}.
func deletePodInvite(c Config, w http.ResponseWriter, r *http.Request) error {
	me, err := podCaller(c, r)
	if err != nil {
		return err
	}
	to, err := podUserID(r.PathValue("user_id"))
	if err != nil {
		return err
	}
	if err := c.Pods.CancelInvite(r.Context(), me, r.PathValue("id"), to); err != nil {
		return podError(c, err, "cancelling a pod invite")
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}

// acceptPodInvite is POST /me/pod-invites/{pod_id}/accept.
func acceptPodInvite(c Config, w http.ResponseWriter, r *http.Request) error {
	me, err := podCaller(c, r)
	if err != nil {
		return err
	}
	s, err := c.Pods.Accept(r.Context(), me, r.PathValue("pod_id"))
	if err != nil {
		return podError(c, err, "accepting a pod invite")
	}
	return writeJSON(w, http.StatusOK, summaryOf(s))
}

// declinePodInvite is POST /me/pod-invites/{pod_id}/decline.
func declinePodInvite(c Config, w http.ResponseWriter, r *http.Request) error {
	me, err := podCaller(c, r)
	if err != nil {
		return err
	}
	if err := c.Pods.Decline(r.Context(), me, r.PathValue("pod_id")); err != nil {
		return podError(c, err, "declining a pod invite")
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}

// postPodLink is POST /pods/{id}/link: mint, or rotate, the invite link.
func postPodLink(c Config, w http.ResponseWriter, r *http.Request) error {
	me, err := podCaller(c, r)
	if err != nil {
		return err
	}
	token, exp, err := c.Pods.MintLink(r.Context(), me, r.PathValue("id"))
	if err != nil {
		return podError(c, err, "making a pod invite link")
	}
	w.Header().Set("Cache-Control", "no-store")
	return writeJSON(w, http.StatusCreated, linkResponse(token, exp))
}

// deletePodLink is DELETE /pods/{id}/link: turn the invite link off.
func deletePodLink(c Config, w http.ResponseWriter, r *http.Request) error {
	me, err := podCaller(c, r)
	if err != nil {
		return err
	}
	if err := c.Pods.ClearLink(r.Context(), me, r.PathValue("id")); err != nil {
		return podError(c, err, "turning off a pod invite link")
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}

// previewPodLink is GET /pods/join/{token}.
func previewPodLink(c Config, w http.ResponseWriter, r *http.Request, token string) error {
	me, err := podCaller(c, r)
	if err != nil {
		return err
	}
	p, err := c.Pods.Preview(r.Context(), me, token)
	if err != nil {
		return podError(c, err, "previewing a pod invite link")
	}
	w.Header().Set("Cache-Control", "no-store")
	return writeJSON(w, http.StatusOK, podPreviewResponse{
		PodID: p.PodID, Name: p.Name, OwnerName: p.OwnerName, MemberCount: p.Members, Member: p.Member,
	})
}

// joinPodByLink is POST /pods/join/{token}.
func joinPodByLink(c Config, w http.ResponseWriter, r *http.Request, token string) error {
	me, err := podCaller(c, r)
	if err != nil {
		return err
	}
	s, err := c.Pods.JoinByLink(r.Context(), me, token)
	if err != nil {
		return podError(c, err, "joining a pod by link")
	}
	return writeJSON(w, http.StatusOK, summaryOf(s))
}

// deletePodMember is DELETE /pods/{id}/members/{user_id}: the owner's
// removal, or leaving (`me`, or your own id). ?new_link=1 on a removal
// also replaces a live invite link and answers it, 200 with the link;
// otherwise 204.
func deletePodMember(c Config, w http.ResponseWriter, r *http.Request) error {
	me, err := podCaller(c, r)
	if err != nil {
		return err
	}
	target := me
	if raw := r.PathValue("user_id"); raw != "me" {
		if target, err = podUserID(raw); err != nil {
			return err
		}
	}
	newLink := r.URL.Query().Get("new_link") == "1" || r.URL.Query().Get("new_link") == "true"
	res, err := c.Pods.RemoveMember(r.Context(), me, r.PathValue("id"), target, newLink)
	if err != nil {
		return podError(c, err, "removing a pod member")
	}
	if res.LinkToken != "" {
		w.Header().Set("Cache-Control", "no-store")
		return writeJSON(w, http.StatusOK, map[string]podLinkResponse{"link": linkResponse(res.LinkToken, res.LinkExpiresAt)})
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}

// adminRenamePod is PUT /admin/pods/{id}/name (ADR 0220 §8).
func adminRenamePod(c Config, w http.ResponseWriter, r *http.Request) error {
	if c.Pods == nil {
		return httpError(http.StatusServiceUnavailable, errPodsOff)
	}
	var body podNameBody
	if err := decodeJSON(w, r, &body); err != nil {
		return err
	}
	id := r.PathValue("id")
	name, err := c.Pods.AdminRename(r.Context(), id, body.Name)
	if err != nil {
		return podError(c, err, "an admin renaming a pod")
	}
	return writeJSON(w, http.StatusOK, map[string]string{"id": id, "name": name})
}

// adminDeletePod is DELETE /admin/pods/{id} (ADR 0220 §8).
func adminDeletePod(c Config, w http.ResponseWriter, r *http.Request) error {
	if c.Pods == nil {
		return httpError(http.StatusServiceUnavailable, errPodsOff)
	}
	if err := c.Pods.AdminDelete(r.Context(), r.PathValue("id")); err != nil {
		return podError(c, err, "an admin deleting a pod")
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}
