function getBenchComparison(host, spec, values, ignore, filter, depth, focus) {
	$.getJSON(host + "/compare", {
		filter:filter,
		spec:spec,
		values:values,
		ignore:ignore
	},
	function(data) {
		if (data["Error"] != null) {
			alert(data["Error"]);
			return;
		}

		var categories = data["Result"];

		if (categories == null) {
			// No result
			return;
		}

		$.each(categories, function(i, group) {
			// A group is a list of benchlists - one per spec value
			var t = makeTable(group, spec);
			$("#compare").append(t);
		});
	});
}

function makeTable(group, spec) {
	var t = $("<table>");
	t.addClass("table");
	t.addClass("table-hover");
	t.addClass("table-striped");
	t.addClass("table-bordered");
	t.addClass("table-nonfluid");

	if (group.length == 0)
		return t;

	var head = $("<thead>");
	t.append(head);
	head.append("<tr><th>" + spec + "</th><th>Time</th></tr>");

	var body = $("<tbody>");
	t.append(body);

	return t;
}
