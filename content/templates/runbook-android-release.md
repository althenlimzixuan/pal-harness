# Android release runbook: <app name>

Copy to `doc/runbooks/android-release.md` before the first Play release, and fill in the
placeholders. The session dry-reads it against the real files before ship (phase skill,
`references/ship.md`, section 0). Steps marked **(you)** are owner-run: the plan lists them under
**Owner-run commands** with their exact text.

- **Package:** `<com.company.product>`: the same as the Play Console app and `app.config.ts`
  `android.package` (the spec's store identity check).
- **Version:** `<x.y.z>` (`version`) and `<n>` (`android.versionCode`, or EAS's auto increment):
  bumped by the phase's version task.
- **EAS project:** `<owner>/<slug>`; build profile `production`.

## 1. Signing credentials (EAS)

1. EAS keeps one Android keystore per app. Let EAS create it on the first production build, and
   keep it on EAS. Run credential set-up from one terminal only: a second terminal set up at the
   same time can create a second keystore.
2. **(you)** Back it up: `eas credentials -p android` → the `production` profile → Keystore →
   Download. This downloads the keystore EAS builds with, not a new one. Store the file and its
   passwords in the password manager, never in the repo.
3. Write its SHA-1 fingerprint (shown by `eas credentials`) here: `<SHA-1>`.

## 2. The first upload locks the upload key

1. With Play App Signing, the first app bundle uploaded to the Play Console registers its signing
   key as the **upload key**. Every later upload must be signed with that same upload key.
2. So the first upload is always an EAS `production` build signed with the keystore from section 1,
   never a local or debug build. Check the fingerprint before uploading.
3. A lost upload key needs an upload key reset through the Play Console, which takes days.

## 3. Build and upload

Under Company mode the upload is held: ship builds from `main`, adds a row to
`doc/releases/held.md`, and stops before step 2 until the one batched release (phase skill,
`references/ship.md`, section 8).

1. **(you)** `eas build --platform android --profile production`
2. **(you)** `eas submit --platform android --latest` (or upload the `.aab` by hand in step 4).

## 4. Play Console: the draft release

1. Testing → Internal testing (or the track the phase names) → Create new release.
2. **Add from library** picks the bundle `eas submit` uploaded; otherwise upload the `.aab`.
3. Release name `<x.y.z>`; release notes from the phase's What to expect now lines.
4. Policy → App content, before the first review:
   - **Advertising ID:** answer whether the app uses an advertising ID. An app that does not must
     not ship the `com.google.android.gms.permission.AD_ID` permission (check the merged manifest);
   - Data safety, content rating, target audience, privacy policy link.
5. Save, review, then **(you)** roll out to the track.

## 5. After the release

1. Write the release in the log: version, version code, track, date.
2. A failed step: the error and what was tried go in the log, and the session sends "needs you".
