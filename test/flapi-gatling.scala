package flapi

import io.gatling.core.Predef._
import io.gatling.http.Predef._
import scala.concurrent.duration._

class test extends Simulation {
  val httpConf = http
    .baseUrl("http://localhost:8000")

  val scn = scenario("FLAPI.test")
    .exec(http("POST:/api/a")
      .post("/api/a"))

  // val scn = scenario("FLAPI.test")
  //   .exec(http("POST:/api/a")
  //     .post("/api/a"))
  //   .pause(100 milliseconds)
  //   .repeat(5) {
  //     exec(http("GET:/api/a")
  //       .get("/api/a"))
  //       .pause(10 milliseconds)
  //   }
  //   .pause(100 milliseconds)
  //   .repeat(2) {
  //     exec(http("GET:/api/b")
  //       .get("/api/b"))
  //   }
  //   .pause(100 milliseconds)
  //   .exec(http("PUT:/api/c")
  //     .put("/api/c"))
  //   .pause(100 milliseconds)
  //   .repeat(2) {
  //     exec(http("GET:/api/c")
  //       .get("/api/c"))
  //   }

  // val scnA = scenario("FLAPI.a")
  //   .exec(http("POST:/api/a")
  //     .post("/api/a"))
  //   .pause(100 milliseconds)
  //   .repeat(5) {
  //     exec(http("GET:/api/a")
  //       .get("/api/a"))
  //       .pause(10 milliseconds)
  //   }

  // val scnB = scenario("FLAPI.b")
  //   .repeat(5) {
  //     exec(http("GET:/api/b")
  //       .get("/api/b"))
  //       .pause(10 milliseconds)
  //   }

  // val scnC = scenario("FLAPI.c")
  //   .exec(http("PUT:/api/c")
  //     .put("/api/c"))
  //   .pause(100 milliseconds)
  //   .repeat(2) {
  //     exec(http("GET:/api/c")
  //       .get("/api/c"))
  //   }

  setUp(
      scn.inject(constantUsersPerSec(100) during(30 minute)).protocols(httpConf)
  .protocols(httpConf))

  // setUp(
  //     scnA.inject(constantUsersPerSec(10) during(10 minute)).protocols(httpConf),
  //     scnB.inject(constantUsersPerSec(10) during(10 minute)).protocols(httpConf),
  //     scnC.inject(constantUsersPerSec(10) during(10 minute)).protocols(httpConf)
  // .protocols(httpConf))

  // setUp(
    // scn.inject(
    // nothingFor(4 seconds),
    // atOnceUsers(10),
    // rampUsers(10) over(5 seconds),
    // rampUsers(100) over(1 minute),
    // constantUsersPerSec(100) during(10 minutes)
    // constantUsersPerSec(20) during(15 seconds) randomized,
    // rampUsersPerSec(10) to 20 during(10 minutes),
    // rampUsersPerSec(10) to 20 during(10 minutes) randomized,
    // splitUsers(1000) into(rampUsers(10) over(10 seconds)) separatedBy(10 seconds),
    // splitUsers(1000) into(rampUsers(10) over(10 seconds)) separatedBy atOnceUsers(30),
    // heavisideUsers(1000) over(10 minutes)
  // .protocols(httpConf))
  // )
}
