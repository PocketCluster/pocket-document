#/usr/bin/env bash

DMG_PATH=${1}
DMG_URL=${2}
NOTE_URL=${3}

if [[ -z ${DMG_URL} ]]; then
    echo "[ERR] empty dmg image url."
    exit
fi

if [[ -z ${DMG_PATH} ]]; then
    echo "[ERR] empty dmg image path"
    exit
fi

# Syntax //DELIMITER//. We use \/ because we need to escape the delimiter.    
DMG_DIRS=(${DMG_PATH//\// })
# Get the length
DIR_LENGTH=${#DMG_DIRS[@]}
# Subtract 1 from the length
LAST_POSITION=$((DIR_LENGTH - 1))
# Get the last position.
LAST_PART=${DMG_DIRS[$LAST_POSITION]}

TITLE=${LAST_PART/.dmg/}
TITLE=${TITLE/-/ }
VERSION=($TITLE)
VERSION=${VERSION[1]}
MODIFIED=$(/usr/local/bin/python -c "import os,time; print time.ctime(os.path.getmtime('${DMG_PATH}'))")
#MODIFIED=$(/usr/bin/stat ${DMG_PATH} | /usr/bin/cut -d' ' -f9 ${FILE_STAT})
FILESIZE=$(/usr/bin/stat ${DMG_PATH} | /usr/bin/cut -d' ' -f8 ${FILE_STAT})
HASH=$(../APPCERT/SparkleSigningTools/sign_update.rb ${DMG_PATH} ../APPCERT/Sparkle/dsa_priv.pem)

echo "<item>
    <title>${TITLE}</title>
    <sparkle:releaseNotesLink>${NOTE_URL}</sparkle:releaseNotesLink>
    <pubDate>${MODIFIED}</pubDate>
    <enclosure
    url=\"${DMG_URL}\" 
    sparkle:version=\"${VERSION}\" 
    length=\"${FILESIZE}\" 
    type=\"application/octet-stream\" 
    sparkle:dsaSignature=\"${HASH}\" />
</item>"