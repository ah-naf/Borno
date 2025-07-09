async function loadExamples() {
  try {
    const res = await fetch("/examples");
    const data = await res.json();
    const select = document.getElementById("example-select");
    select.innerHTML = '<option value="">--Select--</option>';
    data.examples.forEach((ex) => {
      const opt = document.createElement("option");
      opt.value = ex.name;
      opt.textContent = ex.name;
      opt.dataset.code = ex.code;
      select.appendChild(opt);
    });
  } catch (err) {
    console.error(err);
  }
}

function setup() {
  loadExamples();
  const select = document.getElementById("example-select");
  const codeInput = document.getElementById("code-input");
  select.addEventListener("change", () => {
    const option = select.options[select.selectedIndex];
    if (option.dataset.code) {
      codeInput.value = option.dataset.code;
    }
  });

  document.getElementById("run-btn").addEventListener("click", async () => {
    const code = codeInput.value;
    const res = await fetch("/run", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ code }),
    });
    const data = await res.json();
    const output = document.getElementById("output");
    output.textContent = data.output || data.error;
  });
}

document.addEventListener("DOMContentLoaded", setup);
