const state = {
  guilds: [],
  guild: null,
  config: null,
  roles: [],
  channels: []
};

const pages = {
  landing: document.getElementById('landing'),
  servers: document.getElementById('servers'),
  dashboard: document.getElementById('dashboard')
};

const els = {
  serverGrid: document.getElementById('serverGrid'),
  guildName: document.getElementById('guildName'),
  ticketCategory: document.getElementById('ticketCategory'),
  logChannel: document.getElementById('logChannel'),
  transcriptChannel: document.getElementById('transcriptChannel'),
  autoCloseMinutes: document.getElementById('autoCloseMinutes'),
  supportRoles: document.getElementById('supportRoles'),
  departments: document.getElementById('departments'),
  blacklist: document.getElementById('blacklist'),
  blacklistUser: document.getElementById('blacklistUser'),
  toast: document.getElementById('toast')
};

document.getElementById('backToServers').addEventListener('click', () => show('servers'));
document.getElementById('saveButton').addEventListener('click', saveConfig);
document.getElementById('addDepartment').addEventListener('click', () => {
  state.config.departments.push({ name: '', roleId: '', emoji: '' });
  renderDepartments();
});
document.getElementById('addBlacklist').addEventListener('click', () => {
  const value = els.blacklistUser.value.trim();
  if (!value || state.config.blacklist.includes(value)) return;
  state.config.blacklist.push(value);
  els.blacklistUser.value = '';
  renderBlacklist();
});

init();

async function init() {
  const me = await api('/api/me', { optional: true });
  if (!me) {
    show('landing');
    return;
  }
  state.guilds = await api('/api/guilds');
  renderServers();
  show('servers');
}

function show(name) {
  Object.values(pages).forEach(page => page.classList.add('hidden'));
  pages[name].classList.remove('hidden');
}

function renderServers() {
  els.serverGrid.innerHTML = '';
  state.guilds.forEach(guild => {
    const card = document.createElement('button');
    card.className = `server-card ${guild.botPresent ? '' : 'disabled'}`;
    card.disabled = !guild.botPresent;
    card.innerHTML = `
      <span class="server-icon">${initials(guild.name)}</span>
      <span>
        <strong>${escapeHtml(guild.name)}</strong>
        <small>${guild.botPresent ? 'Manage settings' : 'Bot not in server'}</small>
      </span>
    `;
    card.addEventListener('click', () => openGuild(guild));
    els.serverGrid.appendChild(card);
  });
}

async function openGuild(guild) {
  state.guild = guild;
  els.guildName.textContent = guild.name;
  const [config, roles, channels] = await Promise.all([
    api(`/api/config/${guild.id}`),
    api(`/api/roles/${guild.id}`),
    api(`/api/channels/${guild.id}`)
  ]);
  state.config = config;
  state.roles = roles.filter(role => role.name !== '@everyone');
  state.channels = channels;
  renderDashboard();
  show('dashboard');
}

function renderDashboard() {
  renderChannelSelect(els.ticketCategory, 4, state.config.ticketCategoryId);
  renderChannelSelect(els.logChannel, 0, state.config.logChannelId);
  renderChannelSelect(els.transcriptChannel, 0, state.config.transcriptChannelId);
  els.autoCloseMinutes.value = state.config.autoCloseMinutes || 1440;
  renderRoleMulti();
  renderDepartments();
  renderBlacklist();
}

function renderChannelSelect(select, type, selected) {
  select.innerHTML = '<option value="">Select a channel</option>';
  state.channels
    .filter(channel => channel.type === type)
    .forEach(channel => {
      select.appendChild(new Option(`# ${channel.name}`, channel.id, false, channel.id === selected));
    });
}

function renderRoleMulti() {
  els.supportRoles.innerHTML = '';
  state.roles.forEach(role => {
    els.supportRoles.appendChild(new Option(role.name, role.id, false, state.config.supportRoles.includes(role.id)));
  });
}

function renderRoleSelect(selected) {
  const select = document.createElement('select');
  select.innerHTML = '<option value="">Select a role</option>';
  state.roles.forEach(role => {
    select.appendChild(new Option(role.name, role.id, false, role.id === selected));
  });
  return select;
}

function renderDepartments() {
  els.departments.innerHTML = '';
  state.config.departments.forEach((department, index) => {
    const row = document.createElement('div');
    row.className = 'department-row';
    const name = document.createElement('input');
    name.placeholder = 'Department name';
    name.value = department.name || '';
    name.addEventListener('input', () => department.name = name.value);
    const role = renderRoleSelect(department.roleId);
    role.addEventListener('change', () => department.roleId = role.value);
    const emoji = document.createElement('input');
    emoji.placeholder = 'Emoji';
    emoji.value = department.emoji || '';
    emoji.addEventListener('input', () => department.emoji = emoji.value);
    const remove = document.createElement('button');
    remove.className = 'button danger';
    remove.textContent = 'Remove';
    remove.addEventListener('click', () => {
      state.config.departments.splice(index, 1);
      renderDepartments();
    });
    row.append(name, role, emoji, remove);
    els.departments.appendChild(row);
  });
}

function renderBlacklist() {
  els.blacklist.innerHTML = '';
  state.config.blacklist.forEach((userId, index) => {
    const row = document.createElement('div');
    row.className = 'list-row';
    row.innerHTML = `<code>${escapeHtml(userId)}</code>`;
    const remove = document.createElement('button');
    remove.className = 'button danger';
    remove.textContent = 'Remove';
    remove.addEventListener('click', () => {
      state.config.blacklist.splice(index, 1);
      renderBlacklist();
    });
    row.appendChild(remove);
    els.blacklist.appendChild(row);
  });
}

async function saveConfig() {
  state.config.ticketCategoryId = els.ticketCategory.value;
  state.config.logChannelId = els.logChannel.value;
  state.config.transcriptChannelId = els.transcriptChannel.value;
  state.config.autoCloseMinutes = Number(els.autoCloseMinutes.value);
  state.config.supportRoles = Array.from(els.supportRoles.selectedOptions).map(option => option.value);
  state.config.departments = state.config.departments.filter(dept => dept.name.trim());
  state.config = await api(`/api/config/${state.guild.id}`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(state.config)
  });
  renderDashboard();
  toast();
}

async function api(path, options = {}) {
  const response = await fetch(path, options);
  if (options.optional && response.status === 401) return null;
  if (!response.ok) throw new Error(await response.text());
  return response.json();
}

function toast() {
  els.toast.classList.remove('hidden');
  setTimeout(() => els.toast.classList.add('hidden'), 2200);
}

function initials(name) {
  return name.split(/\s+/).slice(0, 2).map(part => part[0] || '').join('').toUpperCase();
}

function escapeHtml(value) {
  return String(value).replace(/[&<>"']/g, char => ({
    '&': '&amp;',
    '<': '&lt;',
    '>': '&gt;',
    '"': '&quot;',
    "'": '&#039;'
  }[char]));
}
