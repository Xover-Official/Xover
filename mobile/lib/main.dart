import 'package:flutter/material.dart';

void main() {
  runApp(const TalosApp());
}

class TalosApp extends StatelessWidget {
  const TalosApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'Talos Mobile',
      theme: ThemeData(
        colorScheme: ColorScheme.fromSeed(seedColor: Colors.cyanAccent),
        useMaterial3: true,
        scaffoldBackgroundColor: Colors.black,
      ),
      home: const DashboardPage(title: 'Talos: Swarm Status'),
    );
  }
}

class DashboardPage extends StatefulWidget {
  const DashboardPage({super.key, required this.title});
  final String title;

  @override
  State<DashboardPage> createState() => _DashboardPageState();
}

class _DashboardPageState extends State<DashboardPage> {
  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: Text(widget.title, style: const TextStyle(color: Colors.white)),
        backgroundColor: Colors.grey[900],
      ),
      body: Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: <Widget>[
            const Icon(
              Icons.hub,
              color: Colors.cyanAccent,
              size: 100,
            ),
            const SizedBox(height: 20),
            const Text(
              'Talos Swarm: ONLINE',
              style: TextStyle(color: Colors.white, fontSize: 24),
            ),
            const SizedBox(height: 10),
            Text(
              'Active Nodes: 12',
              style: TextStyle(color: Colors.grey[400]),
            ),
          ],
        ),
      ),
      floatingActionButton: FloatingActionButton(
        onPressed: () {},
        tooltip: 'Emergency Stop',
        backgroundColor: Colors.red,
        child: const Icon(Icons.stop),
      ),
    );
  }
}
