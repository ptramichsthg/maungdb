async function runQuery() {
    // 1. Sesuaikan ID dengan HTML (queryInput)
    const inputEl = document.getElementById("queryInput");
    if (!inputEl) {
        console.error("Error: Element #queryInput tidak ditemukan.");
        return;
    }
    const q = inputEl.value;

    // 2. Sesuaikan ID Output dengan HTML (queryOutput)
    const out = document.getElementById("queryOutput");
    out.innerHTML = "Processing..."; // Feedback visual

    const res = await API.post("/query", { query: q });

    if (!res.success) {
      // Gunakan style alert-error yang sudah ada di CSS baru
      out.innerHTML = `<div class="alert alert-error">${res.error}</div>`;
      return;
    }

    const data = res.data;
    if (!data || !data.Rows || data.Rows.length === 0) {
      out.innerHTML = `<div class="alert alert-success">${res.Message || "Query OK, tapi tidak ada data (Empty Set)"}</div>`;
      return;
    }

    // Render Table
    let html = "<div class='result-container'><table><thead><tr>";
    data.Columns.forEach(c => html += `<th>${c}</th>`);
    html += "</tr></thead><tbody>";

    data.Rows.forEach(r => {
      html += "<tr>";
      r.forEach(v => html += `<td>${v}</td>`);
      html += "</tr>";
    });

    html += "</tbody></table></div>";
    
    // Tambahkan info jumlah baris
    html += `<div style="margin-top:10px; font-size:0.85rem; color:#666;">Total: ${data.Rows.length} baris</div>`;

    out.innerHTML = html;
}