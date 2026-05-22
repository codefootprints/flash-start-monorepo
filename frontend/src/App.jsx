import { useEffect, useState } from "react";
import "./App.css";

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
	// State untuk aset
	const [resources, setResources] = useState([]);
	const [resourceFormData, setResourceFormData] = useState({
		name: "",
		category: "",
	});

	// State untuk user
	const [users, setUsers] = useState([]);
	const [userFormData, setUserFormData] = useState({
		username: "",
		email: "",
	});

  // State untuk tasks
  const [tasks, setTasks] = useState([])

  const [taskHistory, setTaskHistory] = useState([])

	// State untuk notifikasi
	const [notification, setNotification] = useState({
		message: "",
		type: "",
	});

	const showNotification = (message, type = "success") => {
		setNotification({ message, type });
		setTimeout(() => setNotification({ message: "", type: "" }), 3000);
	};

	const [assignment, setAssignment] = useState({
		user_id: "",
		resource_id: "",
		title: "",
	});

	// Mengambil data dari backend Golang
	const fetchUsers = () => {
		fetch("http://localhost:3000/api/users")
			.then((res) => res.json())
			.then((data) => setUsers(data))
			.catch((err) => console.error("Gagal mengambil data", err));
	};

	const fetchResources = () => {
		fetch("http://localhost:3000/api/resources")
			.then((res) => res.json())
			.then((data) => setResources(data))
			.catch((err) => console.error("Gagal mengambil data", err));
	};

  const fetchTasks = () => {
    fetch("http://localhost:3000/api/tasks")
      .then((res) => res.json())
      .then((data) => setTasks(data))
      .catch((err) => console.error("Gagal mengambil data task", err))
  }

  const fetchTaskHistory = () => {
    fetch("http://localhost:3000/api/tasks/history")
      .then(res => res.json())
      .then(data => setTaskHistory(data))
      .catch(err => console.error("Gagal mengambil riwayat", err))
  }

	const handleUserSubmit = (e) => {
		e.preventDefault();
		fetch("http://localhost:3000/api/users", {
			method: "POST",
			headers: {
				"Content-Type": "application/json",
			},
			body: JSON.stringify(userFormData),
		})
			.then((res) => res.json())
			.then(() => {
				// Reset form
				setUserFormData({
					username: "",
					email: "",
				});
				fetchUsers();
				showNotification("User berhasil ditambahkan!");
			})
			.catch((err) => {
				console.error("Gagal menambah user:", err);
				showNotification("Gagal menambah user", "error");
			});
	};

	const handleResourceSubmit = (e) => {
		e.preventDefault();
		fetch("http://localhost:3000/api/resources", {
			method: "POST",
			headers: {
				"Content-Type": "application/json",
			},
			body: JSON.stringify(resourceFormData),
		})
			.then(async (res) => {
				const data = await res.json()
				if (!res.ok) {
					throw new Error(data.error || "Gagal menambah aset")
				}
				return data
			})
			.then(() => {
				// Reset form
				setResourceFormData({
					name: "",
					category: "",
				});
				// Refresh list tabel
				fetchResources();
				showNotification("Aset berhasil ditambahkan!");
			})
			.catch((err) => {
				showNotification("Gagal menambah aset", "error");
			});
	};

	const handleResourceDelete = (id, name) => {
		if (window.confirm(`Apakah Anda yakin ingin menghapus aset ini?\n${name}`)) {
			fetch(`http://localhost:3000/api/resources/${id}`, {
				method: "DELETE",
			})
				.then(async res => {
					const data = await res.json()
					if (!res.ok) {
						throw new Error(data.error || "Gagal menghapus aset");
					}
					return data
				})
				.then(() => {
					fetchResources();
					showNotification("Aset berhasil dihapus")
				})
				.catch(err => {
					showNotification(err.message, "error")
				});
		}
	};

	const handleAssign = (e) => {
		e.preventDefault();
		fetch("http://localhost:3000/api/tasks", {
			method: "POST",
			headers: {
				"Content-Type": "application/json",
			},
			body: JSON.stringify({
				user_id: parseInt(assignment.user_id),
				resource_id: parseInt(assignment.resource_id),
				title: assignment.title,
				description: "Penugasan otomatis dari dashboard",
			}),
		})
			.then((res) => res.json())
			.then((data) => {
				if (data.error) {
					console.error("Gagal menambah task:", data.error);
				} else {
					fetchResources();
          fetchTasks();
					setAssignment({
						user_id: "",
						resource_id: "",
						title: "",
					});
					showNotification("Penugasan berhasil dibuat!");
				}
			});
	};

  const handleTaskComplete = (id) => {
    fetch(`http://localhost:3000/api/tasks/${id}/complete`, {
      method: "PATCH",
    })
      .then(res => res.json())
      .then(data => {
        if (data.error) {
          showNotification(data.error, "error")
        } else {
          fetchResources()
          fetchTasks()
          fetchTaskHistory()
          showNotification("Aset berhasil dikembalikan!")
        }
      })
      .catch(err => {
        console.error("Gagal mengembalikan aset:", err)
        showNotification("Gagal mengembalikan aset", "error")
      })
  }

	useEffect(() => {
		fetchResources()
    fetchTasks()
		fetchUsers()
    fetchTaskHistory()
	}, []);

	return (
		<>
			{notification.message && (
				<div className={`fixed top-5 right-5 px-6 py-3 rounded-lg shadow-lg border transition-all duration-500 ${notification.type === "success" ? "bg-white border-olive-500 text-olive-500" : "bg-red-50 border-red-500 text-red-600"} z-50`}>
					<p className="font-medium text-sm">{notification.message}</p>
				</div>
			)}

			<div className="min-h-screen bg-stone-50 p-8">
				<div className="max-w-4xl mx-auto">
					<header className="flex justify-between items-center mb-8">
						<h1 className="text-3xl font-bold text-stone-800">START Dashboard</h1>
						<div className="px-4 py-1 bg-olive-100 text-olive-800 rounded-full text-sm font-medium border border-olive-800">{resources.length} Resource Terdata</div>
					</header>

					{/* Form Tambah User */}
					<AccordionSection title="Registrasi Anggota Tim">
						<form onSubmit={handleUserSubmit} className="flex gap-4 flex-wrap">
							<input
								type="text"
								placeholder="Username"
								className="flex-1 p-2 border border-stone-300 rounded focus:outline-none focus:ring-2 focus:ring-olive-500"
								value={userFormData.username}
								onChange={(e) =>
									setUserFormData({
										...userFormData,
										username: e.target.value,
									})
								}
								required
							/>
							<input
								type="email"
								placeholder="Email"
								className="flex-1 p-2 border border-stone-300 rounded focus:outline-none focus:ring-2 focus:ring-olive-500"
								value={userFormData.email}
								onChange={(e) =>
									setUserFormData({
										...userFormData,
										email: e.target.value,
									})
								}
								required
							/>
							<button type="submit" className="px-6 py-2 bg-olive-800 text-white rounded hover:bg-olive-500 font-medium">
								Tambah User
							</button>
						</form>
					</AccordionSection>

					{/* Form Tambah Aset */}
					<AccordionSection title="Tambah Aset Baru">
						<form onSubmit={handleResourceSubmit} className="flex gap-4 flex-wrap">
							<input 
                type="text" 
                placeholder="Nama Aset (Contoh: Macbook Air)" 
                className="flex-1 p-2 border border-stone-300 rounded focus:outline-none focus:ring-2 focus:ring-olive-500" 
                value={resourceFormData.name} 
                onChange={(e) => setResourceFormData({ ...resourceFormData, name: e.target.value })} 
                required 
              />
							<input 
                type="text" 
                placeholder="Kategori" 
                className="flex-1 p-2 border border-stone-300 rounded focus:outline-none focus:ring-2 focus:ring-olive-500" 
                value={resourceFormData.category} 
                onChange={(e) => setResourceFormData({ ...resourceFormData, category: e.target.value })} 
                required 
              />
							<button type="submit" className="px-6 py-2 bg-olive-800 text-white rounded hover:bg-olive-500 font-medium">
								Tambah Aset
							</button>
						</form>
					</AccordionSection>

					{/* Form Assignment */}
					<AccordionSection title="Assign Resource">
						<form onSubmit={handleAssign} className="grid grid-cols-1 md:grid-cols-4 gap-4">
							<input
								type="text"
								placeholder="Judul Tugas"
								className="p-2 border border-stone-300 rounded"
								value={assignment.title}
								onChange={(e) =>
									setAssignment({
										...assignment,
										title: e.target.value,
									})
								}
								required
							/>
							<SelectCustom 
                placeholder="Pilih User" 
                value={assignment.user_id} 
                options={users} 
                onChange={(e) => setAssignment({ ...assignment, user_id: e.target.value })} 
                required 
              />
							<SelectCustom 
                placeholder="Pilih Aset" 
                value={assignment.resource_id} 
                options={resources.filter((r) => r.status === "available")} 
                onChange={(e) => setAssignment({ ...assignment, resource_id: e.target.value })} 
                required 
              />
							<button type="submit" className="px-6 py-2 bg-olive-800 text-white rounded hover:bg-olive-500 font-medium">
								Assign
							</button>
						</form>
					</AccordionSection>

					{/* Tabel Resource */}
					<AccordionSection title="Daftar Resource" defaultOpen={true}>
						<div className="overflow-x-auto">
							<table className="w-full text-left">
								<thead className="bg-stone-100 border-b border-stone-200">
									<tr>
										<th className="px-6 py-4 text-sm font-semibold text-stone-600">Nama Aset</th>
										<th className="px-6 py-4 text-sm font-semibold text-stone-600">Kategori</th>
										<th className="px-6 py-4 text-sm font-semibold text-stone-600">Status</th>
										<th className="px-6 text-sm font-semibold text-stone-600 text-center">Aksi</th>
									</tr>
								</thead>
								<tbody className="divide-y divide-stone-100">
									{resources.map((item) => (
										<tr className="hover:bg-stone-50 transition-colors" key={item.id}>
											<td className="px-6 py-4 text-stone-800 font-medium">{item.name}</td>
											<td className="px-6 py-4 text-stone-500">{item.category}</td>
											<td className="px-6 py-4">
												<span className={`px-3 py-1 rounded-full text-xs font-medium ${item.status === "available" ? "bg-green-100 text-green-700" : "bg-amber-100 text-amber-700"}`}>{item.status}</span>
											</td>
											<td className="px-6 py-4 text-center">
												<button onClick={() => handleResourceDelete(item.id, item.name)} className="text-red-600 hover:text-white hover:bg-red-600 font-medium text-sm rounded px-2 py-1 hover:cursor-pointer">
													Hapus
												</button>
											</td>
										</tr>
									))}
								</tbody>
							</table>
						</div>
					</AccordionSection>

          {/* Tabel Task / Peminjaman Aktif */}
          <AccordionSection title="Daftar Peminjaman Aktif" defaultOpen={true}>
            <div className="overflow-x-auto">
              <table className="w-full text-left">
                <thead className="bg-stone-100 border-b border-stone-200">
                  <tr>
                    <th className="px-6 py-4 text-sm font-semibold text-stone-600">User</th>
                    <th className="px-6 py-4 text-sm font-semibold text-stone-600">Aset</th>
                    <th className="px-6 py-4 text-sm font-semibold text-stone-600">Judul Tugas</th>
                    <th className="px-6 py-4 text-sm font-semibold text-stone-600">Aksi</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-stone-100">
                  {tasks.length > 0 ? (
                    tasks.map(task => (
                      <tr className="hover:bg-stone-50 transition-colors" key={task.id}>
                        <td className="px-6 py-4 text-stone-800">{task.user?.username || "N/A"}</td>
                        <td className="px-6 py-4 text-stone-800 font-medium">{task.resource?.name || "N/A"}</td>
                        <td className="px-6 py-4 text-stone-800 italic">{task.title}</td>
                        <td className="px-6 py-4 text-center">
                          <button
                            onClick={() => handleTaskComplete(task.id)}
                            className="border border-olive-500 text-olive-800 hover:bg-olive-500 hover:text-white font-medium text-sm rounded px-3 py-1 transition-all hover:cursor-pointer focus:outline-none focus:ring-2 focus:ring-olive-500"
                          >
                            Kembalikan
                          </button>
                        </td>
                      </tr>
                    ))
                  ) : (
                    <tr>
                      <td colSpan="4" className="px-4 py-8 text-center text-stone-400">
                        Tidak ada peminjaman aktif
                      </td>
                    </tr>
                  )}
                </tbody>
              </table>
            </div>
          </AccordionSection>

          <AccordionSection title="Riwayat Pengembalian" defaultOpen={true}>
            <div className="overflow-x-auto">
                <table className="w-full text-left">
                    <thead className="bg-stone-100 border-b border-stone-200">
                        <tr>
                            <th className="px-6 py-4 text-sm font-semibold text-stone-600">User</th>
                            <th className="px-6 py-4 text-sm font-semibold text-stone-600">Aset</th>
                            <th className="px-6 py-4 text-sm font-semibold text-stone-600">Selesai Pada</th>
                        </tr>
                    </thead>
                    <tbody className="divide-y divide-stone-100">
                        {taskHistory.length > 0 ? (
                            taskHistory.map((history) => (
                                <tr className="bg-stone-50/30" key={history.id}>
                                    <td className="px-6 py-4 text-stone-600">{history.user?.username || "N/A"}</td>
                                    <td className="px-6 py-4 text-stone-600 font-medium">{history.resource?.name || "N/A"}</td>
                                    <td className="px-6 py-4 text-stone-400 text-xs italic">
                                        {new Date(history.updated_at).toLocaleString("id-ID")}
                                    </td>
                                </tr>
                            ))
                        ) : (
                            <tr>
                                <td colSpan="3" className="px-6 py-8 text-center text-stone-400">
                                    Belum ada riwayat pengembalian
                                </td>
                            </tr>
                        )}
                    </tbody>
                </table>
            </div>
          </AccordionSection>
				</div>
			</div>
		</>
	);
}

export default App;
