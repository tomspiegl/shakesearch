const Controller = {
  search: (ev) => {
    ev.preventDefault();
    const form = document.getElementById("form");
    const data = Object.fromEntries(new FormData(form));
    const response = fetch(`/search?q=${data.query}`).then((response) => {
      response.json().then((results) => {
        Controller.updateTable(results);
      });
    });
  },

  updateTable: (results) => {
    const table = document.getElementById("search-results");
    const rows = [];
    for (let result of results) {
	let text = result.text
	rows.push(`<div>`)
      	rows.push(`<p><b>Title:&nbsp;${result.title}</b></p>`);
      	rows.push(`<p>${text}</p>`);
	rows.push(`</div>`)
    }
    table.innerHTML = rows.join("\n");
  },
};

const form = document.getElementById("form");
form.addEventListener("submit", Controller.search);
