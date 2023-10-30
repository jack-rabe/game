<script lang="ts">
	import { onMount } from 'svelte';

	const url = '://localhost:3333';
	let players: string[] = [];
	let user: string;
	let submitted = false;
	let nameInput: HTMLInputElement;
	let socket: WebSocket;
	let socketId: string;

	const joinLobby = () => {
		if (!user) {
			return;
		}
		const body = JSON.stringify({ Name: user, Id: socketId });
		fetch(`http${url}/addUser`, {
			method: 'POST',
			body,
			headers: {
				'Content-Type': 'application/json'
			}
		});
		nameInput.value = '';
		submitted = true;
	};

	const connectWebSocket = () => {
		socket = new WebSocket(`ws${url}/ws`);

		socket.onopen = (event) => {
			console.log('WebSocket connection opened', event);
			socket.send('hi');
		};
		socket.onmessage = (event) => {
			const data = JSON.parse(event.data);
			if (data.type === 'id') {
				socketId = data.id;
			} else if (data.type === 'join') {
				players.push(data.name);
				players = players;
				if (data.gameReady) {
					window.location.href = `/game/${data.gameId}?name=${user}`;
				}
			} else if (data.type === 'leave') {
				players = players.filter((p) => p !== data.name);
			}
			console.log(data);
		};
		socket.onclose = (event) => {
			alert('Lost connection to server!');
			console.log('WebSocket connection closed', event);
		};
	};

	onMount(() => {
		fetch(`http${url}/getLobbyUsers`)
			.then((res) => res.json())
			.then((res) => (players = res));
		connectWebSocket();
	});
</script>

<h1 class="flex h-14 items-center justify-center bg-secondary text-2xl font-bold text-black">
	Lobby
</h1>
<div class="flex h-2/5 items-center justify-center">
	<div
		class="grid h-32 w-1/2 grid-cols-4 items-center overflow-hidden text-ellipsis rounded-lg bg-secondary hover:overflow-visible"
	>
		{#each players as player}
			<div
				class={`m-3 flex h-5/6 items-center justify-center rounded-lg bg-neutral text-xl font-bold ${
					player === user && submitted ? 'animate-pulse' : ''
				}`}
			>
				{player}
			</div>
		{/each}
	</div>
</div>
<div class="m-2 flex h-2/5 flex-col items-center">
	{#if !submitted}
		<div class="m-4">
			<label for="name" class="label font-bold">Name</label>
			<input
				type="text"
				name="name"
				bind:value={user}
				disabled={submitted}
				bind:this={nameInput}
				class="input border-primary"
			/>
		</div>
		<button class="btn" on:click={joinLobby}>Play</button>
	{/if}
</div>
