Protokol je založený na WebSocket a připojuje se na adresu `ws://server:8000/`.

### Inicializace
Po připojení se potřeba se přihlásit posláním zprávy na server obsahující přihlašovací token. První zpráva, která přijde ze serveru konfigurace hry:

```json
{
	"turns_to_flamout": 1, // jak dlouho žije oheň
	"turns_to_replenish_used_bomb": 12, // za jak dlouho se přičte bomba
	"turns_to_explode": 10, // za jak dlouho bomba exploduje
	"points_per_wall": 1, // kolik je bodů za zničení jedné zdi
	"points_per_kill": 10, // kolik je bodů za zabití hráče
	"points_per_suicide": -10 // kolik je bodů za zabití sám sebe
}
```

Pak vám server bude posílat každé kolo stav hry, to obsahuje hlavně hrací pole, kde se nacházíte, kolik máte bomb a podobně. Tady je ukázka jedné zprávy:

```json
{
	"Name": "Player 1",
	"X": 0, // vaše pozice
	"Y": 0,
	"LastX": 0, // vaše poslední pozice
	"LastY": 0,
	"Bombs": 0, // kolik máte bomb
	"MaxBomb": 3, // velikost inventáře na bomby
	"Radius": 3, // jak daleko vaše bomby zabíjí
	"Alive": false, // jestli žijete
	"Points": 0, // kolik máte bodů
	"Turn": 55, // číslo tahu
	"Board": [
		"#########",
		"#P      #",
		"# # # # #",
		"#       #",
		"# # # # #",
		"#       #",
		"# # # # #",
		"#P     P#",
		"#########"
	],
	// ostatní hráči a jejich parametry
	"Players": [
		{
			"Name": "Player 1",
			"X": 1,
			"Y": 1,
			"LastX": -1,
			"LastY": -1,
			"Bombs": 3,
			"MaxBomb": 3,
			"Radius": 3,
			"Alive": true,
			"Points": 0,
		},
		{
			"Name": "Player 2",
			"X": 7,
			"Y": 1,
			"LastX": -1,
			"LastY": -1,
			"Bombs": 3,
			"MaxBomb": 3,
			"Radius": 3,
			"Alive": true,
			"Points": 0,
		},
		{
			"Name": "Player 3",
			"X": 1,
			"Y": 4,
			"LastX": 1,
			"LastY": 4,
			"Bombs": 3,
			"MaxBomb": 3,
			"Radius": 3,
			"Alive": false,
			"Points": -10,
		},
		{
			"Name": "Player 4",
			"X": 7,
			"Y": 7,
			"LastX": -1,
			"LastY": -1,
			"Bombs": 3,
			"MaxBomb": 3,
			"Radius": 3,
			"Alive": true,
			"Points": 0,
		}
	],
	"Message": "" // občas vám chce server něco říct
}
```

V `Board` je pole textových řetězců a každý z nich reprezentuje jeden sloupec hracího plánu, každý znak reprezentuje jedno políčko podle následující tabulky:

Znak | Význam
---|-----------------
`#` - heškříž | Nezničitelná skála
`.` - tečka | Zničitelná zeď
` ` - mezera  | Volná země
`P`  | Jiný hráč
`B`  | Bomba
`F`  | Oheň způsobený bombou
`n`  | Power up - zvětší inventář na bomby
`r`  | Power up - zvětší dosah bomby