const parseChangelog = require('changelog-parser');

const version = process.argv[3].replace(/^v/,"");
const changelog = process.argv[2];
parseChangelog({
  text: changelog,
}).then(function (result) {
  const versionChangelog = result.versions.find(v => {
    return (v.version || "").replace(/^v/,"") === version;
  }) || {body: ""};

  console.log(versionChangelog.body);
  process.exit(0);
}).catch(function (err) {
  console.log(err);
  process.exit(1);
});
