# Sparkle Update

**How to generate Appcast `version.xml`**

```sh
./updater.sh \
# local file path
~/temp/DMG/PocketCluster-0.1.4.dmg \
# remote file url
https://github.com/stkim1/pocketcluster/releases/download/0.1.4-rc0/PocketCluster-0.1.4.dmg \
# bundle version
3
```

## Issues (04/01/2017)

- As of now, Sparkle's behavior is unexpected most of time. Plus there is no way to force users to upgrade. We need their permission to do so. Though Sparkle has stronger security model.
- Therefore, [like Electron did](https://github.com/electron/electron/pull/162), we need to move to [Squirrel](https://github.com/Squirrel) as it enforce users to update to latest version and it has window version with same api nuance.

## API to consider
- We need to be notified by updater so we can safely terminate on-going daemons.
- We need to prevent users to start application when there is a new update.
- We need to check slave nodes once OSX application is updated and force their agent to be upgraded. **Golang + SCP**

## TODO
- [ ] Setup Sparkle roleout CI (Don't do it manually other than writing release note!)
