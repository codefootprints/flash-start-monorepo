import { useEffect, useState } from "react";

// URL Terpusat Menuju API Gateway (Tasl 11.2)
const GATEWAY_URL = "http://localhost:8000/api/v1"
const GATEWAY_WS = "ws://localhost:8000/api/v1/flash/ws/notifications"

const SelectCustom = ({ label, value, onChange, options, placeholder, required = false }) => {
	return (
		<div className="relative flex-1">
			{/* Label Opsional */}
			{label && <label className="block text-sm font-medium text-stone-600 mb-1">{label}</label>}

			<div className="relative">
				<select value={value} onChange={onChange} required={required} className="w-full p-2 pr-10 border border-stone-300 rounded focus:outline-none focus:ring-2 focus:ring-olive-500 appearance-none bg-white transition-all text-stone-700">
					<option value="">{placeholder || "Pilih opsi"}</option>
					{options.map((opt) => (
						<option key={opt.id} value={opt.id}>
							{opt.username || opt.name || opt.label}
						</option>
					))}
				</select>

				{/* Ikon Panah Manual (Pengganti panah browser) */}
				<div className="pointer-events-none absolute inset-y-0 right-0 flex items-center px-2 text-stone-400">
					<svg className="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path d="M19 9l-7 7-7-7" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" />
					</svg>
				</div>
			</div>
		</div>
	);
};

const AccordionSection = ({ title, children, defaultOpen = false }) => {
	const [isOpen, setIsOpen] = useState(defaultOpen);

	return (
		<div className="bg-white rounded-xl shadow-sm border border-stone-200 mb-4 overflow-hidden">
			<button onClick={() => setIsOpen(!isOpen)} className="w-full flex justify-between items-center p-6 hover:bg-stone-50 transition-colors focus:outline-none">
				<h2 className="text-lg font-semibold text-stone-700">{title}</h2>
				<svg className={`w-6 h-6 text-stone-400 transition-transform duration-300 ${isOpen ? "rotate-180" : ""}`} fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
				</svg>
			</button>

			<div className={`transition-all duration-500 ease-in-out overflow-hidden ${isOpen ? "max-h-[1000px] opacity-100" : "max-h-0 opacity-0"}`}>
				<div className="p-6 border-t border-stone-100">{children}</div>
			</div>
		</div>
	);
};

function App() {
	// --- STATE CORE GLOBAL ---
	const [currentView, setCurrentView] = useState("billing") // "billing" atau "assets" (Task 11.1)
	const [token, setToken] = useState(localStorage.getItem("token") || "")
	const [notification, setNotification] = useState({ message: "", type: "" })
	
	// --- STATE UNTUK FLOW UTAMA START (ASSETS) ---
	const [assets, setAssets] = useState([])
	const [authForm, setAuthForm] = useState({ username: "", password: "" })
	const [isRegistering, setIsRegistering] = useState(false)

	// --- STATE UNTUK FLOW UTAMA FLASH (BILLING) ---
	const [transactions, setTransactions] = useState([])
	const [expandedTxId, setExpandedTxId] = useState(null)

	const showNotification = (message, type = "success") => {
		setNotification({ message, type })
		setTimeout(() => {
			setNotification({ message: "", type: "" })
		}, 4000);
	}

	// --- EFFECT 1: LIVE STREAM WEBSOCKET VIA GATEWAY ---
	useEffect(() => {
		const ws = new WebSocket(GATEWAY_WS)

		ws.onopen = () => {
			console.log("🔌 Connected to Consolidated API Gateway WS Proxy")
			showNotification("Koneksi Live Stram FLASH Aktif via Gateway!")
		}

		ws.onmessage = (event) => {
			const data = JSON.parse(event.data)
			if (data.event === "NEW_TRANSACTION") {
				const newTx = JSON.parse(data.payload)
				setTransactions(prev => [newTx, ...prev])
				showNotification(`🎯 Transaksi Baru: ${newTx.invoice_number} - Rp ${newTx.total_amount.toLocaleString("id-ID")}`, "info")
			}
		}
		return () => ws.close()
	}, [])

	// --- EFFECT 2: FETCH DATA HISTORI BILLING ---
	useEffect(() => {
		fetch(`${GATEWAY_URL}/flash/api/transactions`)
		.then(res => res.json())
		.then(data => {
			if (Array.isArray(data)) setTransactions(data)
		})
		.catch(err => console.error("Gagal memuat history transaksi:", err))
	})

	// --- EFFECT 3: FETCH DATA ASET (DIPROTEKSI JWT GATEWAY) ---
	const fetchAssets = () => {
		if (!token) return
		fetch(`${GATEWAY_URL}/start/assets`, {
			headers: { "Authorization": `Bearer ${token}` }
		})
			.then(res => {
				if (res.status === 401) throw new Error("Sesi kedaluwarsa")
					return res.json()
			})
			.then(data => {
				if (Array.isArray(data)) setAssets(data)
			})
		.catch(err => {
			handleLogout()
			showNotification("Sesi login habis, silakan masuk kembali", "error")
		})
	}

	useEffect(() => {
		if (token && currentView === "assets") fetchAssets()
	}, [token, currentView])

	// --- HANDLER AUTENTIKASI (START SERVICE VIA GATEWAY) ---
	const handleAuthSubmit = (e) => {
		e.preventDefault()
		const endpoint = isRegistering ? "register" : "login"

		fetch(`${GATEWAY_URL}/start/${endpoint}`, {
			method: "POST",
			headers: { "Content-Type": "application/json" },
			body: JSON.stringify(authForm)
		})
			.then(res => res.json())
			.then(data => {
				if (data.token) {
					localStorage.setItem("token", data.token)
					setToken(data.token)
					showNotification("Selamat datang kembali! Autentikasi sukses")
					setAuthForm({ username: "", password: "" })
				} else if (isRegistering) {
					showNotification("Registrasi berhasil, silakan masuk!")
					setIsRegistering(false)
				} else {
					showNotification(data.error || "Gagal memproses autentikasi", "error")
				}
			})
			.catch(() => showNotification("Gagal menghubungi Security Gate Gateway", "error"))
	}

	const handleLogout = () => {
		localStorage.removeItem("token")
		setToken("")
		setAssets([])
	}

	return (
    <div className="flex min-h-screen bg-stone-100 font-sans">
      {/* Toast Notification Banner */}
      {notification.message && (
        <div className={`fixed top-5 right-5 px-6 py-3 rounded-xl shadow-md border transition-all duration-500 ${notification.type === "success" ? "bg-white border-green-500 text-green-600" : notification.type === "info" ? "bg-olive-50 border-olive-600 text-olive-900" : "bg-red-50 border-red-400 text-red-600"} z-50`}>
          <p className="font-semibold text-xs tracking-wide">{notification.message}</p>
        </div>
      )}

      {/* 🧭 NAVIGATION SIDEBAR PANEL (Task 11.1) */}
      <aside className="w-64 bg-stone-900 text-stone-200 flex flex-col border-r border-stone-800">
        <div className="p-6 border-b border-stone-800">
          <h2 className="text-xl font-black tracking-wider text-white">MONOREPO OS</h2>
          <p className="text-stone-500 text-xs italic mt-0.5">Consolidated Enterprise Hub</p>
        </div>

        <nav className="flex-1 p-4 space-y-1.5">
          <button onClick={() => setCurrentView("billing")} className={`w-full flex items-center space-x-3 px-4 py-3 rounded-lg text-sm font-medium transition-colors ${currentView === "billing" ? "bg-olive-800 text-white font-bold" : "hover:bg-stone-800 text-stone-400"}`}>
            <span>⚡ FLASH Billing Feed</span>
          </button>
          <button onClick={() => setCurrentView("assets")} className={`w-full flex items-center space-x-3 px-4 py-3 rounded-lg text-sm font-medium transition-colors ${currentView === "assets" ? "bg-olive-800 text-white font-bold" : "hover:bg-stone-800 text-stone-400"}`}>
            <span>📦 START Asset Depot</span>
          </button>
        </nav>

        {token && (
          <div className="p-4 border-t border-stone-800 bg-stone-950/40">
            <button onClick={handleLogout} className="w-full text-left px-4 py-2 text-xs font-semibold text-red-400 hover:text-red-300 transition-colors">
              ➔ Keluar Sistem
            </button>
          </div>
        )}
      </aside>

      {/* MAIN VIEW CONTROLLER WORKSPACE */}
      <main className="flex-1 p-10 overflow-y-auto">
        {/* VIEW A: FLASH LIVE BILLING FEED */}
        {currentView === "billing" && (
          <div className="max-w-4xl mx-auto">
            <header className="flex justify-between items-center mb-8 pb-5 border-b border-stone-200">
              <div>
                <h1 className="text-2xl font-black text-stone-800">FLASH Engine</h1>
                <p className="text-stone-500 text-xs italic mt-0.5">Asynchronous Signal Telemetry Stream</p>
              </div>
              <div className="px-4 py-1 bg-olive-100 text-olive-800 rounded-full text-xs font-bold border border-olive-700 animate-pulse">🔴 GATEWAY TUNNEL LIVE</div>
            </header>

            <AccordionSection title="Live Transaction Logs" defaultOpen={true}>
              <div className="overflow-x-auto">
                <table className="w-full text-left table-fixed">
                  <thead className="bg-stone-50 border-b border-stone-200">
                    <tr>
                      <th className="w-1/12 px-4 py-3.5 text-xs font-bold text-stone-500 text-center"></th>
                      <th className="w-3/12 px-4 py-3.5 text-xs font-bold text-stone-500">Invoice</th>
                      <th className="w-3/12 px-4 py-3.5 text-xs font-bold text-stone-500">Operator</th>
                      <th className="w-3/12 px-4 py-3.5 text-xs font-bold text-stone-500">Total</th>
                      <th className="w-2/12 px-4 py-3.5 text-xs font-bold text-stone-500 text-right">Timestamp</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-stone-100">
                    {transactions.length > 0 ? (
                      transactions.map((tx) => (
                        <tr key={tx.id} className="p-0">
                          <td colSpan="5" className="p-0">
                            <div onClick={() => setExpandedTxId(expandedTxId === tx.id ? null : tx.id)} className="flex items-center w-full cursor-pointer py-4 hover:bg-stone-50 transition-colors">
                              <div className="w-1/12 text-center flex justify-center">
                                <svg className={`w-3.5 h-3.5 text-stone-400 transition-transform duration-300 ${expandedTxId === tx.id ? "rotate-180" : ""}`} fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2.5} d="M19 9l-7 7-7-7" />
                                </svg>
                              </div>
                              <div className="w-3/12 px-4 text-olive-800 font-bold tracking-wider text-sm">{tx.invoice_number}</div>
                              <div className="w-3/12 px-4 text-stone-600 font-medium text-sm">{tx.cashier_name}</div>
                              <div className="w-3/12 px-4 text-stone-900 font-bold text-sm">Rp {tx.total_amount.toLocaleString("id-ID")}</div>
                              <div className="w-2/12 px-4 text-right text-stone-400 text-xs italic">{new Date(tx.created_at).toLocaleTimeString("id-ID")}</div>
                            </div>

                            {/* Collapsible Details */}
                            <div className={`overflow-hidden transition-all duration-300 bg-stone-50/50 ${expandedTxId === tx.id ? "max-h-[500px] p-4 border-t border-b border-stone-100" : "max-h-0"}`}>
                              <div className="bg-white rounded-lg border border-stone-200 overflow-hidden shadow-2xs">
                                <table className="w-full text-xs text-left">
                                  <thead className="bg-stone-50 border-b border-stone-100">
                                    <tr>
                                      <th className="px-4 py-2 font-bold text-stone-500">Nama Barang</th>
                                      <th className="px-4 py-2 font-bold text-stone-500 text-center">Qty</th>
                                      <th className="px-4 py-2 font-bold text-stone-500 text-right">Subtotal</th>
                                    </tr>
                                  </thead>
                                  <tbody className="divide-y divide-stone-100">
                                    {tx.orders?.map((item) => (
                                      <tr key={item.id}>
                                        <td className="px-4 py-2.5 font-medium text-stone-700">{item.product?.name || `Product ID ${item.product_id}`}</td>
                                        <td className="px-4 py-2.5 text-center font-bold text-stone-600">{item.quantity}x</td>
                                        <td className="px-4 py-2.5 text-right font-bold text-stone-800">Rp {item.sub_total.toLocaleString("id-ID")}</td>
                                      </tr>
                                    ))}
                                  </tbody>
                                </table>
                              </div>
                            </div>
                          </td>
                        </tr>
                      ))
                    ) : (
                      <tr>
                        <td colSpan="5" className="px-4 py-8 text-center text-stone-400 text-xs italic">Menunggu pancaran data dari gateway...</td>
                      </tr>
                    )}
                  </tbody>
                </table>
              </div>
            </AccordionSection>
          </div>
        )}

        {/* VIEW B: START ASSET DEPOT (PROTECTED REGION) */}
        {currentView === "assets" && (
          <div className="max-w-4xl mx-auto">
            <header className="flex justify-between items-center mb-8 pb-5 border-b border-stone-200">
              <div>
                <h1 className="text-2xl font-black text-stone-800">START Depot</h1>
                <p className="text-stone-500 text-xs italic mt-0.5">Corporate Inventory Vault & Protection Area</p>
              </div>
            </header>

            {!token ? (
              /* Auth Form Box */
              <div className="max-w-md mx-auto bg-white rounded-2xl border border-stone-200 shadow-xs p-8 mt-10">
                <h3 className="text-lg font-bold text-stone-800 mb-1">{isRegistering ? "Buat Akun Terpusat" : "Proteksi Security Gate"}</h3>
                <p className="text-stone-400 text-xs mb-6">Silakan autentikasi kredensial Anda untuk membuka proxy rute internal database aset.</p>
                <form onSubmit={handleAuthSubmit} className="space-y-4">
                  <div>
                    <label className="block text-xs font-bold text-stone-500 uppercase tracking-wider mb-1.5">Username</label>
                    <input type="text" required value={authForm.username} onChange={(e) => setAuthForm({...authForm, username: e.target.value})} className="w-full bg-stone-50 border border-stone-200 rounded-xl px-4 py-3 text-sm focus:outline-none focus:border-stone-400 transition-colors" />
                  </div>
                  <div>
                    <label className="block text-xs font-bold text-stone-500 uppercase tracking-wider mb-1.5">Password</label>
                    <input type="password" required value={authForm.password} onChange={(e) => setAuthForm({...authForm, password: e.target.value})} className="w-full bg-stone-50 border border-stone-200 rounded-xl px-4 py-3 text-sm focus:outline-none focus:border-stone-400 transition-colors" />
                  </div>
                  <button type="submit" className="w-full bg-stone-900 text-white text-sm font-bold py-3.5 rounded-xl hover:bg-stone-800 transition-colors shadow-xs">
                    {isRegistering ? "Daftar Akun Baru" : "Buka Kunci Akses"}
                  </button>
                </form>
                <button onClick={() => setIsRegistering(!isRegistering)} className="w-full text-center text-xs text-stone-400 font-medium mt-4 hover:underline">
                  {isRegistering ? "Sudah punya akun? Login di sini" : "Belum terdaftar? Buat akun korporat baru"}
                </button>
              </div>
            ) : (
              /* Secured Asset Table View */
              <div className="bg-white rounded-2xl border border-stone-200 shadow-2xs p-6">
                <div className="flex justify-between items-center mb-6">
                  <h3 className="font-bold text-stone-800 text-lg">Daftar Logistik Aset Perusahaan</h3>
                  <span className="px-3 py-1 bg-green-50 text-green-700 text-[10px] font-black rounded-md uppercase tracking-wider border border-green-200">🛡️ JWT Verified Guarded</span>
                </div>
                <div className="overflow-x-auto">
                  <table className="w-full text-left">
                    <thead className="bg-stone-50 border-b border-stone-200 text-stone-500 text-xs font-bold">
                      <tr>
                        <th className="px-4 py-3">Nama Aset</th>
                        <th className="px-4 py-3">Kategori</th>
                        <th className="px-4 py-3 text-center">Kuantitas</th>
                      </tr>
                    </thead>
                    <tbody className="divide-y divide-stone-100 text-sm text-stone-600 font-medium">
                      {assets.length > 0 ? (
                        assets.map((asset) => (
                          <tr key={asset.id} className="hover:bg-stone-50/40 transition-colors">
                            <td className="px-4 py-3.5 text-stone-900 font-bold">{asset.name}</td>
                            <td className="px-4 py-3.5"><span className="px-2 py-0.5 bg-stone-100 border border-stone-200 text-stone-500 rounded text-xs">{asset.category}</span></td>
                            <td className="px-4 py-3.5 text-center font-bold text-stone-800">{asset.quantity} unit</td>
                          </tr>
                        ))
                      ) : (
                        <tr>
                          <td colSpan="3" className="px-4 py-8 text-center text-stone-400 text-xs italic">Data logistik kosong atau Anda belum menambahkan aset apa pun.</td>
                        </tr>
                      )}
                    </tbody>
                  </table>
                </div>
              </div>
            )}
          </div>
        )}
      </main>
    </div>
	)
}

export default App;
