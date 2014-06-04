/* Package confy is a simple JSON configuration loader.

Lets jump right into an example, say you have the following
configuration file (in JSON) for a simple http server with
access to a database.

	{
		"address": ":8050",
		"db": {
			"host": "localhost",
			"port": 5000,
			"user": "admin",
			"passwd": "hackme",
		}
	}

A typical way to load this kind of configuration is to make a global
struct that can hold the configuration. So lets do that.

	package config

	type Config struct {
		Address string
		Database DatabaseConfig `json:"db"`
	}

	type DatabaseConfig struct {
		Host string
		Port int
		User string
		Passwd string
	}

	// decode with json.NewDecoder(file).Decode(&c)
	//
	// use as c.Database.Host, c.Database.Port, etc...

Issues

This approach works, and is fairly simple to understand. But there are some
issues with it.

1. Most of this configuration isn't global, it's intended for some specific part of your
	program. The database config above is only needed by your sql.DB setup.
2. Handling default values in your configuration quickly gets ugly.
3. If you want to write the config back to file you need to remove default values again,
	because you don't want all your default values in the file.

Basic Congo

Congo handles the points above for you. You can re-use existing types except this time
they don't have to be in the same package (or file) or even know about each others existance.

	// we continue with the types above, except we change the Config type to
	// since it now doesn't need to act as a root anymore.
	type ServerConf struct {
		Address string
	}

	// db.go, some arbitrary file that handles your database setup
	var databaseConf DatabaseConfig

	func init() {
		// Here we add a child to the congo.Default tree, you can
		// use your own created congo.Config if you wish to.
		//
		// We name the child "db" as that is the field name used
		// in the JSON, and use the databaseConf variable as
		// target for this field.
		congo.AddSub("db", &databaseConf)
	}

	// server.go, some arbitrary file that handles your http server setup
	var serverConf ServerConf

	func init() {
		// Since ServeConf is technically still the root object in the
		// JSON, we set it as the root of the congo.Default tree.
		congo.SetRoot(&serverConf)
	}

Now we've setup two independant configurations, that will receive their information
from a single file. To load a configuration we can do the following in our main.

	func main() {
		if err := congo.LoadFile("configuration.json"); err != nil {
			panic("failed to load configuration file:" + err.Error())
		}
	}

This will have populated both serverConf and databaseConf correctly. This handles
issue #1 very explicitly, as a bonus congo gave you #3 in the background already,
even though we have no default handling yet.

Defaults

Congo uses interface Defaulter to handle defaults. If a type is added as a root or
child and implements Defaulter, the Default method will be called before unmarshalling
occurs.

	func (s *ServerConf) Default() {
		s.Address = "localhost:8050"
	}

And that's all that is needed. It is critical that Default methods are deterministic,
because congo will call Default multiple times throughout the parsing process. This gives
you an easy fix to issue 2, and aids in the solution of issue 3 because we now have
a deterministic way of getting default values we can compare to when marshalling.

Issues with Congo

Congo isn't perfect either though, there are various issues with the current implementation.

	1. Only JSON objects are allowed to be roots, especially the exclusion of arrays
		is something painful for some use cases. This can be fixed though!
	2. Slow-ish, underneath there are a lot of round-trips between JSON and Go, this
		might be fixable.
	3. Concurrent-safety, congo currently does not try to be safe when concurrently
		accessed. Suggestions on how much safety is needed is welcome.
	3.1 This also means reloading configurations is currently not-safe even though
		you can reload by repeating the loading step.

See method and function documentation for more info on each.
*/
package congo
