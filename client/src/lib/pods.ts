// pods.ts — pods: groups of players (ADR 0220 §10 and §11, slice 3 of
// #4880): the wire types, the calls, and the words each row and button
// says.
//
// Mirrors server/internal/lobby/pods_http.go. Hand-maintained, like
// friends.ts; update both sides in lockstep. Times are Unix milliseconds.

import { LobbyApiError, authFetch } from "./session";

export const PODS_HASH = "#/pods";

export type PodRole = "owner" | "member";

/** One row of GET /me/pods `pods`, and what accept, join and POST /pods answer. */
export interface PodRow {
  id: string;
  name: string;
  member_count: number;
  role: PodRole;
}

export interface PodInviteForMe {
  pod_id: string;
  pod_name: string;
  invited_by: { display_name: string; avatar_url?: string };
  since: number;
}

export interface MyPods {
  pods: PodRow[];
  invites: PodInviteForMe[];
  max_pods: number;
  max_members: number;
}

export interface PodMember {
  user_id: string;
  display_name: string;
  avatar_url?: string;
  joined_at?: number;
  sharing?: boolean;
  owner?: boolean;
  /** The owner's own list keeps a member they blocked as one row (§6). */
  blocked?: boolean;
}

export interface PodPendingInvite {
  user_id?: string;
  display_name: string;
  invited_by?: { display_name: string };
  since: number;
  mine?: boolean;
  can_cancel?: boolean;
}

export interface PodLinkState {
  live: boolean;
  expires_at?: number;
}

export interface PodDetail {
  id: string;
  name: string;
  role: PodRole;
  owner?: { user_id: string; display_name: string; avatar_url?: string };
  member_count: number;
  max_members: number;
  members: PodMember[];
  invites: PodPendingInvite[];
  /** The owner's only. */
  link?: PodLinkState;
}

/** What minting a link answers; the token is shown this once. */
export interface PodLink {
  token: string;
  path: string;
  expires_at: number;
}

export interface JoinPreview {
  name: string;
  owner_name: string;
  member_count: number;
  member: boolean;
  pod_id?: string;
}

/** The one answer every dead link gets (ADR 0220 §3b). */
export const DEAD_LINK = "This invite link can’t be used.";

/** The page that opens a pod. */
export function podHash(id: string): string {
  return `#/pods/${encodeURIComponent(id)}`;
}

/** The full URL to post, from a minted link's `path`. */
export function podLinkURL(path: string, origin: string = location.origin): string {
  return `${origin}/${path}`;
}

const enc = encodeURIComponent;

export async function fetchMyPods(): Promise<MyPods> {
  const res = await authFetch("/me/pods");
  const b = (await res.json()) as Partial<MyPods>;
  return {
    pods: b.pods ?? [],
    invites: b.invites ?? [],
    max_pods: b.max_pods ?? 10,
    max_members: b.max_members ?? 12,
  };
}

export async function createPod(name: string): Promise<PodRow> {
  const res = await authFetch("/pods", { method: "POST", body: JSON.stringify({ name }) });
  return (await res.json()) as PodRow;
}

export async function fetchPod(id: string): Promise<PodDetail> {
  const res = await authFetch(`/pods/${enc(id)}`);
  const b = (await res.json()) as PodDetail;
  return { ...b, members: b.members ?? [], invites: b.invites ?? [] };
}

export async function renamePod(id: string, name: string): Promise<void> {
  await authFetch(`/pods/${enc(id)}`, { method: "PATCH", body: JSON.stringify({ name }) });
}

export async function deletePod(id: string): Promise<void> {
  await authFetch(`/pods/${enc(id)}`, { method: "DELETE" });
}

export async function handOverPod(id: string, userID: string): Promise<void> {
  await authFetch(`/pods/${enc(id)}/owner`, {
    method: "PUT",
    body: JSON.stringify({ user_id: userID }),
  });
}

export async function invitePodFriend(id: string, userID: string): Promise<void> {
  await authFetch(`/pods/${enc(id)}/invites`, {
    method: "POST",
    body: JSON.stringify({ user_id: userID }),
  });
}

export async function cancelPodInvite(id: string, userID: string): Promise<void> {
  await authFetch(`/pods/${enc(id)}/invites/${enc(userID)}`, { method: "DELETE" });
}

export async function answerPodInvite(podID: string, accept: boolean): Promise<void> {
  await authFetch(`/me/pod-invites/${enc(podID)}/${accept ? "accept" : "decline"}`, {
    method: "POST",
  });
}

export async function mintPodLink(id: string): Promise<PodLink> {
  const res = await authFetch(`/pods/${enc(id)}/link`, { method: "POST" });
  return (await res.json()) as PodLink;
}

export async function turnOffPodLink(id: string): Promise<void> {
  await authFetch(`/pods/${enc(id)}/link`, { method: "DELETE" });
}

/** Leave (`me`) or, as the owner, remove a member; `newLink` also rotates a live link. */
export async function removePodMember(
  id: string,
  userID: string,
  newLink = false,
): Promise<PodLink | null> {
  const res = await authFetch(
    `/pods/${enc(id)}/members/${enc(userID)}${newLink ? "?new_link=1" : ""}`,
    { method: "DELETE" },
  );
  if (res.status === 200) {
    const b = (await res.json()) as { link?: PodLink };
    return b.link ?? null;
  }
  return null;
}

export async function previewPodLink(token: string): Promise<JoinPreview> {
  const res = await authFetch(`/pods/join/${enc(token)}`);
  return (await res.json()) as JoinPreview;
}

export async function joinPodByLink(token: string): Promise<PodRow> {
  const res = await authFetch(`/pods/join/${enc(token)}`, { method: "POST" });
  return (await res.json()) as PodRow;
}

// --- words -------------------------------------------------------------

/** podErrorLine is what a page says when a call failed. */
export function podErrorLine(err: unknown): string {
  if (err instanceof LobbyApiError) {
    if (err.status === 503) return "Pods aren’t available on this server.";
    if (err.status === 404) return "That pod isn’t there any more.";
    const m = err.message.trim();
    if (m) return m.charAt(0).toUpperCase() + m.slice(1) + (/[.!?]$/.test(m) ? "" : ".");
  }
  return "That didn’t work. Try again.";
}

export function memberCountLine(n: number): string {
  return n === 1 ? "1 member" : `${n} members`;
}

function dateText(ms: number): string {
  return new Date(ms).toLocaleDateString(undefined, {
    year: "numeric",
    month: "short",
    day: "numeric",
  });
}

export function joinedLine(at: number | undefined): string {
  return at && at > 0 ? `joined ${dateText(at)}` : "";
}

/** linkLine is what the Invite link card says about the link's state. */
export function linkLine(link: PodLinkState | undefined): string {
  if (!link?.live) return "No invite link. Make one to post in Discord.";
  return link.expires_at ? `Works until ${dateText(link.expires_at)}.` : "The link is live.";
}

/** pickableFriends drops friends already in the pod or invited, by name. */
export function pickableFriends<T extends { user_id: string; display_name: string }>(
  friends: T[],
  pod: Pick<PodDetail, "members" | "invites">,
): T[] {
  const taken = new Set<string>(pod.members.map((m) => m.user_id));
  for (const i of pod.invites) if (i.user_id) taken.add(i.user_id);
  return friends
    .filter((f) => !taken.has(f.user_id))
    .sort((a, b) => a.display_name.localeCompare(b.display_name));
}
