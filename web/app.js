const moviesEl = document.getElementById("movies");
const bookingTitle = document.getElementById("booking-title");
const bookingSubtitle = document.getElementById("booking-subtitle");
const seatGridEl = document.getElementById("seat-grid");
const scheduleRowEl = document.getElementById("schedule-row");
const hallLabelEl = document.getElementById("hall-label");
const timeLabelEl = document.getElementById("time-label");
const ticketsListEl = document.getElementById("tickets-list");
const summaryCaptionEl = document.getElementById("summary-caption");
const totalPriceEl = document.getElementById("total-price");
const btnBook = document.getElementById("btn-book");
const btnClear = document.getElementById("btn-clear");

const ticketTypes = {
  adult:  { label: "Взрослый", price: 2500 },
  student:{ label: "Студент", price: 1900 },
  child:  { label: "Детский", price: 1600 }
};

const seatConfig = { rows: 8, cols: 12 };
let movies = [];
let selectedMovie = null;
let selectedShow = null; // { hall, time }
const selectedSeats = new Map(); // key -> { rowLabel, seatNumber, type }

function rowLabel(index) {
  return String.fromCharCode("A".charCodeAt(0) + index);
}

function generateScheduleForMovie(movie, index) {
  // Просто детерминированный "рандом": 3 сеанса, залы 1–8
  const baseHour = 11 + (index % 3) * 3;
  const hall1 = (index % 8) + 1;
  const hall2 = ((index + 3) % 8) + 1;
  const hall3 = ((index + 5) % 8) + 1;
  const time1 = String(baseHour).padStart(2, "0") + ":00";
  const time2 = String(baseHour + 3).padStart(2, "0") + ":30";
  const time3 = String(baseHour + 6).padStart(2, "0") + ":15";
  return [
    { hall: hall1, time: time1 },
    { hall: hall2, time: time2 },
    { hall: hall3, time: time3 }
  ];
}

async function loadMovies() {
  moviesEl.innerHTML = "<p style='color:var(--muted);font-size:13px;'>Загружаем афишу...</p>";
  try {
    const res = await fetch("/api/movies");
    if (!res.ok) throw new Error("HTTP " + res.status);
    const data = await res.json();
    if (!Array.isArray(data) || data.length === 0) {
      moviesEl.innerHTML = "<p style='color:var(--muted);font-size:13px;'>Фильмы пока не добавлены.</p>";
      return;
    }
    movies = data;
    renderMovies();
  } catch (e) {
    moviesEl.innerHTML = "<p style='color:var(--muted);font-size:13px;'>Не удалось загрузить фильмы.</p>";
  }
}

function renderMovies() {
  moviesEl.innerHTML = "";
  movies.forEach((m, index) => {
    const card = document.createElement("div");
    card.className = "movie-card";
    card.dataset.id = m.id;

    const poster = document.createElement("div");
    poster.className = "movie-poster";
    if (m.posterUrl) {
      poster.style.backgroundImage = "url('" + m.posterUrl.replace(/'/g, "\\'") + "')";
      poster.innerHTML = "";
    } else {
      const initial = (m.title || "?").trim().charAt(0).toUpperCase();
      poster.innerHTML = "<span>" + initial + "</span>";
    }

    const info = document.createElement("div");
    info.className = "movie-info";
    const h3 = document.createElement("h3");
    h3.textContent = m.title || "(Без названия)";
    const meta = document.createElement("div");
    meta.className = "movie-meta";
    meta.textContent =
      (m.genre || "Жанр не указан") +
      " · " +
      (m.duration ? (m.duration + " мин") : "длительность неизвестна");

    const tags = document.createElement("div");
    tags.className = "movie-tags";

    const tagRating = document.createElement("span");
    tagRating.className = "tag tag-rating";
    tagRating.textContent = "Рейтинг: " + (m.rating ?? "-");
    tags.appendChild(tagRating);

    const tagId = document.createElement("span");
    tagId.className = "tag";
    tagId.textContent = "ID #" + m.id;
    tags.appendChild(tagId);

    info.appendChild(h3);
    if (m.description) {
      const p = document.createElement("p");
      p.style.margin = "2px 0 0";
      p.style.fontSize = "12px";
      p.style.color = "var(--muted)";
      p.textContent = m.description;
      info.appendChild(p);
    }
    info.appendChild(meta);
    info.appendChild(tags);

    card.appendChild(poster);
    card.appendChild(info);

    card.addEventListener("click", () => selectMovie(m, index, card));

    moviesEl.appendChild(card);
  });
}

function selectMovie(movie, index, cardEl) {
  selectedMovie = movie;
  selectedShow = null;
  selectedSeats.clear();
  updateSummary();
  hallLabelEl.textContent = "–";
  timeLabelEl.textContent = "Время не выбрано";
  scheduleRowEl.innerHTML = "";
  seatGridEl.innerHTML = "";

  document.querySelectorAll(".movie-card").forEach(c => c.classList.remove("active"));
  if (cardEl) cardEl.classList.add("active");

  bookingTitle.textContent = movie.title || "Выбранный фильм";
  bookingSubtitle.textContent =
    (movie.genre || "Жанр не указан") +
    " · " +
    (movie.duration ? (movie.duration + " мин") : "длительность неизвестна") +
    (movie.rating ? " · Рейтинг: " + movie.rating : "");

  const schedule = generateScheduleForMovie(movie, index);
  schedule.forEach((s, i) => {
    const chip = document.createElement("button");
    chip.type = "button";
    chip.className = "chip";
    chip.innerHTML = "<span>" + s.time + " · Зал " + s.hall + "</span>";
    chip.addEventListener("click", () => {
      document.querySelectorAll(".chip").forEach(c => c.classList.remove("active"));
      chip.classList.add("active");
      selectedShow = s;
      hallLabelEl.textContent = "Зал " + s.hall;
      timeLabelEl.textContent = "Сегодня, " + s.time;
      selectedSeats.clear();
      renderSeatGrid();
      updateSummary();
    });
    if (i === 0) {
      chip.click();
    }
    scheduleRowEl.appendChild(chip);
  });
}

function renderSeatGrid() {
  seatGridEl.innerHTML = "";
  if (!selectedShow) {
    seatGridEl.innerHTML = "<p style='color:var(--muted);font-size:12px;text-align:center;margin-top:12px;'>Сначала выберите время и зал.</p>";
    return;
  }
  const frag = document.createDocumentFragment();
  for (let r = 0; r < seatConfig.rows; r++) {
    const row = document.createElement("div");
    row.className = "seat-row";
    const label = document.createElement("div");
    label.className = "seat-row-label";
    const rLabel = rowLabel(r);
    label.textContent = rLabel;
    row.appendChild(label);

    for (let c = 1; c <= seatConfig.cols; c++) {
      const key = rLabel + c;
      const seat = document.createElement("button");
      seat.type = "button";
      seat.className = "seat";
      seat.dataset.key = key;
      seat.innerHTML = "<span></span>";
      seat.addEventListener("click", () => toggleSeat(key, rLabel, c, seat));
      row.appendChild(seat);
    }
    frag.appendChild(row);
  }
  seatGridEl.appendChild(frag);
}

function toggleSeat(key, rowLabelValue, seatNumber, seatEl) {
  if (!selectedShow) return;
  if (selectedSeats.has(key)) {
    selectedSeats.delete(key);
    seatEl.classList.remove("selected");
  } else {
    selectedSeats.set(key, {
      rowLabel: rowLabelValue,
      seatNumber: seatNumber,
      type: "adult"
    });
    seatEl.classList.add("selected");
  }
  updateSummary();
}

function updateSummary() {
  ticketsListEl.innerHTML = "";
  if (selectedSeats.size === 0 || !selectedShow || !selectedMovie) {
    summaryCaptionEl.textContent = "Выберите один или несколько мест на схеме.";
    totalPriceEl.textContent = "0 ₸";
    btnBook.disabled = true;
    return;
  }
  summaryCaptionEl.textContent =
    "Зал " + selectedShow.hall + ", " + selectedShow.time + " · " + selectedSeats.size + " мест(а)";

  selectedSeats.forEach((ticket, key) => {
    const container = document.createElement("div");
    container.className = "ticket-item";

    const place = document.createElement("div");
    place.innerHTML = "Ряд " + ticket.rowLabel + ", место " + ticket.seatNumber +
      "<small>Билет " + key + "</small>";

    const typeSel = document.createElement("select");
    typeSel.className = "ticket-type";
    for (const [code, t] of Object.entries(ticketTypes)) {
      const opt = document.createElement("option");
      opt.value = code;
      opt.textContent = t.label;
      if (code === ticket.type) opt.selected = true;
      typeSel.appendChild(opt);
    }
    typeSel.addEventListener("change", () => {
      ticket.type = typeSel.value;
      renderSummary();
    });

    const price = document.createElement("div");
    price.className = "ticket-price";
    price.dataset.key = key;

    container.appendChild(place);
    container.appendChild(typeSel);
    container.appendChild(price);

    ticketsListEl.appendChild(container);
  });

  function renderSummary() {
    let sum = 0;
    selectedSeats.forEach((ticket, key) => {
      const type = ticketTypes[ticket.type] || ticketTypes.adult;
      const itemPrice = type.price;
      sum += itemPrice;
      const priceEl = ticketsListEl.querySelector('.ticket-price[data-key="' + key + '"]');
      if (priceEl) {
        priceEl.textContent = itemPrice.toLocaleString("ru-RU") + " ₸";
      }
    });
    totalPriceEl.textContent = sum.toLocaleString("ru-RU") + " ₸";
    btnBook.disabled = selectedSeats.size === 0;
  }
  renderSummary();
}

btnClear.addEventListener("click", () => {
  selectedSeats.clear();
  document.querySelectorAll(".seat.selected").forEach(s => s.classList.remove("selected"));
  updateSummary();
});

btnBook.addEventListener("click", () => {
  if (!selectedMovie || !selectedShow || selectedSeats.size === 0) return;
  alert(
    "Демо: бронь создана.\n\nФильм: " + selectedMovie.title +
    "\nЗал: " + selectedShow.hall +
    "\nВремя: " + selectedShow.time +
    "\nМест: " + selectedSeats.size +
    "\nСумма: " + totalPriceEl.textContent +
    "\n\n(В учебной версии данные не сохраняются в базе.)"
  );
});

loadMovies();

