function openList() {
	$("#listBtn").addClass("active");
	$("#compareBtn").removeClass("active");
}

function openCompare() {
	$("#listBtn").removeClass("active");
	$("#compareBtn").addClass("active");
}

var hash = window.location.hash.substring(1);
if (hash == "compare")
	openCompare()
