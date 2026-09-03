const jamming = "Your device is being jammed" 
for (let i = 0; i < data.vas_events.data.length; i++) { 

	if (data.vas_events.data[i].vas === "SPEED" ) {
		if (data.vas_events.data[i].data) {
			data.vas_events.data[i].vas = "Speed violation, allowed speed is " + data.vas_events.data[i].data.speed_limit + "km/h your speed is: " + data.vas_events.data[i].data.realtime.speed + "km/h"
			data.vas_events.data[i].data.description="Speed violation"
		}
		else if (data.vas_events.data[i].message) {
			data.vas_events.data[i].vas = "Speed violation, allowed speed is: " + data.vas_events.data[i].message.speed_limit + " km/h your speed is: " + data.vas_events.data[i].message.realtime.speed + "km/h"
			data.vas_events.data[i].message="Speed violation"
		}
	}
	else if (data.vas_events.data[i].vas === "JAMMING" ) {
		data.vas_events.data[i].vas = jamming	
	}
	else if (data.vas_events.data[i].vas === "GEOFENCE" ) {
		data.vas_events.data[i].vas = data.vas_events.data[i].vas + " : " + data.vas_events.data[i].data.geofence + " was violated"
	}

}
