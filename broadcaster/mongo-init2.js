init = false;
if (!db.isMaster().ismaster) {
  print("mongo-init2.js - Error: primary not ready, initialize ...")
  rs.initiate();
  quit(1);
} else {
  if (!init) {
    admin = db.getSiblingDB("admin");

    admin.createUser(
      {
        user: "test",
        pwd: "pass",
        roles: ["readWriteAnyDatabase"]
      }
    );
    init = true;
  }
}
