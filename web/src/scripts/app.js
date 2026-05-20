/**
 * Entry point for frontend script bundling.
 * Pulls in SCSS which Webpack extracts to app.css.
 */
import '../styles/main.scss';

/**
 * DOM element references used by the interactive scanning interface.
 */
const overlay = document.getElementById('overlay');
const overlayMsg = document.getElementById('overlay-msg');
const btnPick = document.getElementById('btn-pick');
const btnScan = document.getElementById('btn-scan');
const pathInput = document.getElementById('path-input');

/**
 * Guard execution to ensure elements exist in the DOM (interactive serve mode only).
 */
if (btnScan && pathInput) {

	/**
	 * Native folder picker interface trigger.
	 * Spawns the OS's native folder dialog from the Go backend.
	 */
	btnPick.addEventListener('click', async () => {
		try {
			overlay.classList.add('active');
			overlayMsg.textContent = 'Opening native folder explorer...';

			const res = await fetch('/api/browse', { method: 'POST' });
			overlay.classList.remove('active');

			if (res.status === 200) {
				const data = await res.json();
				if (data.path) {
					pathInput.value = data.path;
				}
			}
		} catch (err) {
			overlay.classList.remove('active');
			console.error('Error selecting folder:', err);
			alert('Failed to open native folder explorer.');
		}
	});

	/**
	 * Triggers the directory scan workflow.
	 */
	function startScan() {
		const p = pathInput.value.trim();
		if (!p) { pathInput.focus(); return; }

		overlay.classList.add('active');
		overlayMsg.textContent = `Scanning ${p} …`;

		const url = new URL(window.location.href);
		url.searchParams.set('path', p);
		window.location.href = url.toString();
	}

	btnScan.addEventListener('click', startScan);
	pathInput.addEventListener('keydown', e => { if (e.key === 'Enter') startScan(); });
}

/**
 * Quarantine API interaction trigger.
 * Moves a selected file to the secure local quarantine directory.
 * @param {HTMLButtonElement} button - The button element that triggered the action.
 */
window.quarantine = async function quarantine(button) {
	const path = button.dataset.path;
	if (!confirm(`Move this file to quarantine?\n\n${path}`)) return;

	button.disabled = true;
	button.textContent = 'Moving…';

	try {
		const res = await fetch('/api/quarantine', {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ path }),
		});

		if (!res.ok) {
			const body = await res.text();
			button.disabled = false;
			button.textContent = 'Quarantine';
			alert(body || 'Unable to quarantine file.');
			return;
		}

		button.textContent = 'Quarantined ✓';
		button.closest('tr').style.opacity = '0.5';
	} catch {
		button.disabled = false;
		button.textContent = 'Quarantine';
		alert('Network error — could not reach the server.');
	}
};
