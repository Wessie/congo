/*
Package confy is a simple JSON configuration loader.

Lets jump right into an example, say you have the following
configuration file (in JSON) for a simple http server with
access to a database.

	{
		"Address": ":8050",
		"DB": {
			"Host": "localhost",
			"Port": 5000,
			"User": "admin",
			"Passwd": "hackme",
		}
	}

A typical way to load this kind of configuration is to make a global
struct that can hold the configuration. So lets do that.

	package config

	type Config struct {
		Address string
		Database DatabaseConfig `json:"DB"`
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

Issues with traditional

This approach works, and is fairly simple to understand. But there are some
issues with it.

	1. Most of this configuration isn't global, it's intended for some specific part of your
		program. The database config above is only needed by your sql.DB setup.
	2. Handling default values in your configuration quickly gets ugly.
	3. If you want to write the config back to file you need to remove default values again,
		because you don't want all your default values in the file. This is mostly
		fixed behind the scenes with congo.

The congo way

Congo handles the points above for you. We can reuse our existing types, except this time
they don't have to be in the same package (or file) or even know about each others existance.

	// we continue with the types above, except we can change the Config type
	// to something more local.
	type ServerConf struct {
		Address string
	}

	// db.go, some arbitrary file that handles your database setup
	var databaseConf DatabaseConfig

	func init() {
		// Here we add a child to the congo.Default tree, you can
		// use your own created congo.Config if you so desire.
		//
		// We name the child "DB" as that is the field name used
		// in the JSON, and use the databaseConf variable as
		// target for this field just as before.
		congo.AddSub("DB", &databaseConf)
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

This will populate both serverConf and databaseConf, returning any errors
that occurred. This handles issue 1 very explicitly (and cleanly).

Handling defaults

Congo uses interface Defaulter to handle defaults. If a type handled by congo implements
Defaulter, the Default will be called before unmarshalling into it.

	func (s *ServerConf) Default() {
		s.Address = "localhost:8050"
	}

And that's all that is needed. It is critical that Default methods are deterministic,
because congo will call Default multiple times throughout the parsing process. This gives
you an simple to understand, and local fix to issue 2.

We haven't mentioned issue 3 yet, because said issue has already been fixed simply by
having default values set. Congo will do the rest behind the scenes to avoid writing
back values that are equal to their defaults.

Issues with congo

Congo isn't perfect either, there are various issues with the current implementation.

	1. Only JSON objects are allowed to be roots, especially the exclusion of arrays
		is something painful for some use cases. This could be fixed in a future
		version of congo, but is currently not high on the priority list.
	2. Slow-ish, underneath there are a lot of round-trips between JSON and Go. This
		isn't a big issue because you generally don't load many configuration
		files in a programs lifetime.
	3. Concurrent-safety, congo currently does not try to be safe when concurrently
		accessed. Suggestions on how much safety is needed is welcome.
	4. With 3 this means reloading configurations is currently not-safe even though
		you can (unsafely) reload by repeating the loading step.
*/
package congo
