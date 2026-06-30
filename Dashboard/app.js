document.addEventListener('DOMContentLoaded', () => {
    
    // View Selectors
    const viewLanding = document.getElementById('view-landing');
    const viewServers = document.getElementById('view-servers');
    const viewDashboard = document.getElementById('view-dashboard');
    
    // Global Elements
    const navDashBtns = document.querySelectorAll('.nav-dash-btn');
    const activeGuildCard = document.querySelector('.active-guild');
    const backToServersBtn = document.getElementById('back-to-servers');
    const logoutBtns = document.querySelectorAll('.logout-btn');
    
    // Drawer & Toast Controls
    const drawer = document.getElementById('settings-drawer');
    const closeDrawerBtn = document.getElementById('close-drawer');
    const drawerTitle = document.getElementById('drawer-title');
    const formContainer = document.getElementById('form-container');
    const applyDrawerBtn = document.getElementById('apply-drawer-btn');
    const globalSaveBtn = document.getElementById('global-save-btn');
    const saveToast = document.getElementById('save-toast');

    // Central Data Model State
    let systemState = {
        roles: "Director, Management, Staff Agent",
        liveCategory: "837492018473920184",
        archiveCategory: "928374910293847561",
        namingFormat: "ticket-{username}"
    };

    // Routing Navigation Flow Handles
    navDashBtns.forEach(btn => {
        btn.addEventListener('click', () => {
            switchView(viewServers);
        });
    });

    activeGuildCard.addEventListener('click', () => {
        switchView(viewDashboard);
    });

    backToServersBtn.addEventListener('click', () => {
        switchView(viewServers);
    });

    logoutBtns.forEach(btn => {
        btn.addEventListener('click', () => {
            switchView(viewLanding);
        });
    });

    function switchView(targetView) {
        [viewLanding, viewServers, viewDashboard].forEach(v => v.classList.add('hidden'));
        targetView.classList.remove('hidden');
    }

    // Grid System Click Handling & Configurations Mapping
    const gridCards = document.querySelectorAll('.grid-card');
    gridCards.forEach(card => {
        card.addEventListener('click', () => {
            if (card.classList.contains('tier-locked')) return;

            const target = card.getAttribute('data-target');
            openConfigurationBlock(target);
        });
    });

    function openConfigurationBlock(blockType) {
        formContainer.innerHTML = ''; // Clean current elements
        drawer.classList.remove('hidden');

        if (blockType === 'modal-general') {
            drawerTitle.innerText = "System Operations Setup";
            formContainer.innerHTML = `
                <div class="form-element">
                    <label>Support Team Roles</label>
                    <input type="text" id="input-roles" value="${systemState.roles}">
                </div>
            `;
        } else if (blockType === 'modal-categories') {
            drawerTitle.innerText = "Routing & Anchors";
            formContainer.innerHTML = `
                <div class="form-element">
                    <label>Active Category ID</label>
                    <input type="text" id="input-live" value="${systemState.liveCategory}">
                </div>
                <div class="form-element">
                    <label>Archive Category ID</label>
                    <input type="text" id="input-archive" value="${systemState.archiveCategory}">
                </div>
            `;
        } else if (blockType === 'modal-tickets') {
            drawerTitle.innerText = "Ticket Formats Layout";
            formContainer.innerHTML = `
                <div class="form-element">
                    <label>Naming Convension String</label>
                    <input type="text" id="input-naming" value="${systemState.namingFormat}">
                </div>
            `;
        } else {
            // Default placeholder dynamic generator for uncoded cards
            drawerTitle.innerText = "Module Details";
            formContainer.innerHTML = `<p style="color: var(--text-secondary)">This administrative structural field configuration option is online and waiting backend hooks.</p>`;
        }
    }

    // Modal Drawer Applier
    applyDrawerBtn.addEventListener('click', () => {
        const inputRoles = document.getElementById('input-roles');
        const inputLive = document.getElementById('input-live');
        const inputArchive = document.getElementById('input-archive');
        const inputNaming = document.getElementById('input-naming');

        if (inputRoles) systemState.roles = inputRoles.value;
        if (inputLive) systemState.liveCategory = inputLive.value;
        if (inputArchive) systemState.archiveCategory = inputArchive.value;
        if (inputNaming) systemState.namingFormat = inputNaming.value;

        drawer.classList.add('hidden');
    });

    closeDrawerBtn.addEventListener('click', () => drawer.classList.add('hidden'));
    document.querySelector('.drawer-overlay').addEventListener('click', () => drawer.classList.add('hidden'));

    // Global Push Save Simulation
    globalSaveBtn.addEventListener('click', () => {
        saveToast.classList.remove('hidden');
        setTimeout(() => {
            saveToast.classList.add('hidden');
        }, 2500);
    });
});